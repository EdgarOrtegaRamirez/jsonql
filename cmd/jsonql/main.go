package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/EdgarOrtegaRamirez/jsonql/internal/filter"
	"github.com/EdgarOrtegaRamirez/jsonql/internal/output"
	"github.com/EdgarOrtegaRamirez/jsonql/internal/query"
)

var (
	format     string
	filterExpr string
	sortBy     string
	sortDesc   bool
	limit      int
	prefix     string
	compact    bool
)

var rootCmd = &cobra.Command{
	Use:   "jsonql",
	Short: "Ergonomic JSON query CLI",
	Long: `jsonql is a lightweight JSON query tool with simple path syntax,
filtering, sorting, and multiple output formats.

Examples:
  jsonql --path name data.json
  jsonql --path 'users[*].name' data.json
  jsonql --path name --filter 'age > 25' data.json
  jsonql --path items --sort price --limit 5 data.json
  jsonql --path data --format jsonl data.json
  cat data.json | jsonql --path items
`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runQuery(args)
	},
}

var queryCmd = &cobra.Command{
	Use:   "query [file...]",
	Short: "Query JSON files (same as root command)",
	Long:  `Extract, filter, sort, and format data from JSON files.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runQuery(args)
	},
}

var prettyCmd = &cobra.Command{
	Use:   "pretty [file...]",
	Short: "Pretty-print JSON",
	Long:  `Format JSON from files or stdin with indentation.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var reader *query.Reader
		var err error

		if len(args) == 0 {
			reader, err = query.NewReader(os.Stdin)
		} else {
			reader, err = query.NewReaderFromFile(args[0])
		}
		if err != nil {
			return fmt.Errorf("read input: %w", err)
		}

		data, err := reader.ReadAll()
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}

		w := output.NewJSONWriter(compact)
		if err := w.Write(data, os.Stdout); err != nil {
			return fmt.Errorf("output: %w", err)
		}
		return nil
	},
}

var infoCmd = &cobra.Command{
	Use:   "info [file...]",
	Short: "Show JSON structure summary",
	Long:  `Display a summary of JSON structure: types, keys, array lengths, nested depth.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var reader *query.Reader
		var err error

		if len(args) == 0 {
			reader, err = query.NewReader(os.Stdin)
		} else {
			reader, err = query.NewReaderFromFile(args[0])
		}
		if err != nil {
			return fmt.Errorf("read input: %w", err)
		}

		data, err := reader.ReadAll()
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}

		w := output.NewInfoWriter()
		if err := w.Write(data, os.Stdout); err != nil {
			return fmt.Errorf("info: %w", err)
		}
		return nil
	},
}

func runQuery(args []string) error {
	var reader *query.Reader
	var err error

	if len(args) == 0 {
		reader, err = query.NewReader(os.Stdin)
	} else {
		reader, err = query.NewReaderFromFile(args[0])
	}
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}

	// Apply query path
	var result interface{}
	if prefix == "" {
		result, err = query.Execute(reader)
	} else {
		result, err = query.ExecutePath(reader, prefix)
	}
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	// Apply filter if specified
	if filterExpr != "" {
		var f *filter.Parser
		f, err = filter.NewParser()
		if err != nil {
			return fmt.Errorf("filter: %w", err)
		}
		result, err = f.Evaluate(result, filterExpr)
		if err != nil {
			return fmt.Errorf("filter: %w", err)
		}
	}

	// Apply sort if specified
	if sortBy != "" {
		result, err = query.SortResults(result, sortBy, sortDesc)
		if err != nil {
			return fmt.Errorf("sort: %w", err)
		}
	}

	// Apply limit
	if limit > 0 {
		result, err = query.LimitResults(result, limit)
		if err != nil {
			return fmt.Errorf("limit: %w", err)
		}
	}

	// Output
	outFormat := output.ParseFormat(format)

	var w output.Writer
	switch outFormat {
	case output.JSON:
		w = output.NewJSONWriter(compact)
	case output.JSONL:
		w = output.NewJSONLWriter()
	case output.TEXT:
		w = output.NewTextWriter()
	default:
		w = output.NewJSONWriter(compact)
	}

	if err := w.Write(result, os.Stdout); err != nil {
		return fmt.Errorf("output: %w", err)
	}

	return nil
}

func init() {
	// Persistent flags on root — available for all subcommands
	rootCmd.PersistentFlags().StringVarP(&prefix, "path", "p", "", "JSONPath-like query (e.g., 'users[*].name')")
	rootCmd.PersistentFlags().StringVarP(&filterExpr, "filter", "f", "", "Filter expression (e.g., 'age > 25')")
	rootCmd.PersistentFlags().StringVarP(&sortBy, "sort", "s", "", "Sort by field name")
	rootCmd.PersistentFlags().BoolVarP(&sortDesc, "sort-desc", "S", false, "Sort descending")
	rootCmd.PersistentFlags().IntVarP(&limit, "limit", "l", 0, "Limit number of results")
	rootCmd.PersistentFlags().StringVarP(&format, "format", "F", "json", "Output format: json, jsonl, text")
	rootCmd.PersistentFlags().BoolVarP(&compact, "compact", "c", false, "Compact output (no pretty print)")
}

func main() {
	rootCmd.AddCommand(queryCmd)
	rootCmd.AddCommand(prettyCmd)
	rootCmd.AddCommand(infoCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
