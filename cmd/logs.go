package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kev/cloudflared-cli/internal/config"
	"github.com/kev/cloudflared-cli/internal/ui"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Tail tunnel logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		env := config.ResolveEnv(projectDir, cfgEnv)
		follow, _ := cmd.Flags().GetBool("follow")
		lines, _ := cmd.Flags().GetInt("lines")

		if !config.IsProjectInitialized(projectDir) {
			return fmt.Errorf("project not initialized, run 'cloudflared-project init' first")
		}

		logPath := filepath.Join(projectDir, "logs", env+".log")

		f, err := os.Open(logPath)
		if err != nil {
			return fmt.Errorf("no logs found for env %q (has the tunnel been deployed with --detach?)", env)
		}
		defer f.Close()

		if lines > 0 {
			printLastLines(f, lines)
		}

		if follow {
			ui.Dim("Following logs for [%s]... (Ctrl+C to stop)", env)
			return followLogs(f)
		}

		if lines == 0 {
			// Default: print last 20 lines
			printLastLines(f, 20)
		}

		return nil
	},
}

func printLastLines(f *os.File, n int) {
	// Read all lines, keep last n
	scanner := bufio.NewScanner(f)
	var allLines []string
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}

	start := 0
	if len(allLines) > n {
		start = len(allLines) - n
	}
	for _, line := range allLines[start:] {
		fmt.Println(line)
	}
}

func followLogs(f *os.File) error {
	// Seek to end
	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		return err
	}

	reader := bufio.NewReader(f)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// Wait briefly and retry
				continue
			}
			return err
		}
		fmt.Print(line)
	}
}

func init() {
	logsCmd.Flags().BoolP("follow", "f", false, "follow log output")
	logsCmd.Flags().IntP("lines", "n", 0, "number of lines to show")
	rootCmd.AddCommand(logsCmd)
}
