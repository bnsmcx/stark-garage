package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"

	"toolbox-memory/internal/memory"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Extract the subcommand before parsing flags.
	subcmd := os.Args[1]

	// Global --db flag: parse it from os.Args[2:] for subcommands that use it.
	// Each subcommand defines its own flag set.
	switch subcmd {
	case "init":
		cmdInit(os.Args[2:])
	case "write":
		cmdWrite(os.Args[2:])
	case "search":
		cmdSearch(os.Args[2:])
	case "read":
		cmdRead(os.Args[2:])
	case "list":
		cmdList(os.Args[2:])
	case "delete":
		cmdDelete(os.Args[2:])
	case "prune":
		cmdPrune(os.Args[2:])
	case "promote":
		cmdPromote(os.Args[2:])
	case "stats":
		cmdStats(os.Args[2:])
	case "version":
		fmt.Println(version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", subcmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `Usage: toolbox-memory <command> [flags]

Commands:
  init      Create the database file
  write     Store a memory entry
  search    Full-text search within a namespace
  read      Get a single entry by namespace+key (increments hit_count)
  list      List active/validated entries in a namespace
  delete    Delete an entry by namespace+key
  prune     Run lifecycle transitions (active->stale->archived)
  promote   Mark an entry as promoted
  stats     Print lifecycle statistics
  version   Print version

Global flag (all commands except version):
  --db PATH   Override database path (default: .claude/memory/toolbox.db)`)
}

// openDB opens the database using the --db flag value, or the default path.
func openDB(dbPath string) (*memory.DB, error) {
	if dbPath != "" {
		return memory.Open(dbPath)
	}
	return memory.OpenDefault()
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func jsonOut(v interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		fatal("json encode error: %v", err)
	}
}

// --- Subcommands ---

func cmdInit(args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	dbPath := fs.String("db", "", "database path")
	fs.Parse(args)

	db, err := openDB(*dbPath)
	if err != nil {
		fatal("init failed: %v", err)
	}
	db.Close()
	fmt.Println("database initialized")
}

func cmdWrite(args []string) {
	fs := flag.NewFlagSet("write", flag.ExitOnError)
	dbPath := fs.String("db", "", "database path")
	ns := fs.String("ns", "", "namespace (required)")
	agent := fs.String("agent", "", "agent name (required)")
	key := fs.String("key", "", "key (required)")
	value := fs.String("value", "", "value (required)")
	fs.Parse(args)

	if *ns == "" || *agent == "" || *key == "" || *value == "" {
		fatal("write requires --ns, --agent, --key, and --value")
	}

	db, err := openDB(*dbPath)
	if err != nil {
		fatal("open db: %v", err)
	}
	defer db.Close()

	id, err := db.Store(*ns, *agent, *key, *value)
	if err != nil {
		fatal("write failed: %v", err)
	}

	jsonOut(map[string]interface{}{
		"id":        id,
		"namespace": *ns,
		"key":       *key,
		"status":    "ok",
	})
}

func cmdSearch(args []string) {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	dbPath := fs.String("db", "", "database path")
	ns := fs.String("ns", "", "namespace (required)")
	query := fs.String("query", "", "search query (required)")
	limitStr := fs.String("limit", "10", "max results")
	fs.Parse(args)

	if *ns == "" || *query == "" {
		fatal("search requires --ns and --query")
	}

	limit, err := strconv.Atoi(*limitStr)
	if err != nil {
		fatal("invalid --limit: %v", err)
	}

	db, err := openDB(*dbPath)
	if err != nil {
		fatal("open db: %v", err)
	}
	defer db.Close()

	entries, err := db.Search(*ns, *query, limit)
	if err != nil {
		fatal("search failed: %v", err)
	}

	enc := json.NewEncoder(os.Stdout)
	for _, e := range entries {
		if err := enc.Encode(e); err != nil {
			fatal("json encode error: %v", err)
		}
	}
	if len(entries) == 0 {
		fmt.Println("[]")
	}
}

func cmdRead(args []string) {
	fs := flag.NewFlagSet("read", flag.ExitOnError)
	dbPath := fs.String("db", "", "database path")
	ns := fs.String("ns", "", "namespace (required)")
	key := fs.String("key", "", "key (required)")
	fs.Parse(args)

	if *ns == "" || *key == "" {
		fatal("read requires --ns and --key")
	}

	db, err := openDB(*dbPath)
	if err != nil {
		fatal("open db: %v", err)
	}
	defer db.Close()

	entry, err := db.Get(*ns, *key)
	if err != nil {
		fatal("read failed: %v", err)
	}

	jsonOut(entry)
}

func cmdList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	dbPath := fs.String("db", "", "database path")
	ns := fs.String("ns", "", "namespace (required)")
	fs.Parse(args)

	if *ns == "" {
		fatal("list requires --ns")
	}

	db, err := openDB(*dbPath)
	if err != nil {
		fatal("open db: %v", err)
	}
	defer db.Close()

	entries, err := db.List(*ns)
	if err != nil {
		fatal("list failed: %v", err)
	}

	enc := json.NewEncoder(os.Stdout)
	for _, e := range entries {
		if err := enc.Encode(e); err != nil {
			fatal("json encode error: %v", err)
		}
	}
	if len(entries) == 0 {
		fmt.Println("[]")
	}
}

func cmdDelete(args []string) {
	fs := flag.NewFlagSet("delete", flag.ExitOnError)
	dbPath := fs.String("db", "", "database path")
	ns := fs.String("ns", "", "namespace (required)")
	key := fs.String("key", "", "key (required)")
	fs.Parse(args)

	if *ns == "" || *key == "" {
		fatal("delete requires --ns and --key")
	}

	db, err := openDB(*dbPath)
	if err != nil {
		fatal("open db: %v", err)
	}
	defer db.Close()

	if err := db.Delete(*ns, *key); err != nil {
		fatal("delete failed: %v", err)
	}

	jsonOut(map[string]string{
		"namespace": *ns,
		"key":       *key,
		"status":    "deleted",
	})
}

func cmdPrune(args []string) {
	fs := flag.NewFlagSet("prune", flag.ExitOnError)
	dbPath := fs.String("db", "", "database path")
	maxActive := fs.Int("max-active", 200, "max active entries before overflow archival")
	fs.Parse(args)

	db, err := openDB(*dbPath)
	if err != nil {
		fatal("open db: %v", err)
	}
	defer db.Close()

	n, err := db.Prune(*maxActive)
	if err != nil {
		fatal("prune failed: %v", err)
	}

	jsonOut(map[string]interface{}{
		"transitions": n,
		"status":      "ok",
	})
}

func cmdPromote(args []string) {
	fs := flag.NewFlagSet("promote", flag.ExitOnError)
	dbPath := fs.String("db", "", "database path")
	ns := fs.String("ns", "", "namespace (required)")
	key := fs.String("key", "", "key (required)")
	to := fs.String("to", "", "promotion target (required)")
	fs.Parse(args)

	if *ns == "" || *key == "" || *to == "" {
		fatal("promote requires --ns, --key, and --to")
	}

	db, err := openDB(*dbPath)
	if err != nil {
		fatal("open db: %v", err)
	}
	defer db.Close()

	if err := db.Promote(*ns, *key, *to); err != nil {
		fatal("promote failed: %v", err)
	}

	jsonOut(map[string]string{
		"namespace": *ns,
		"key":       *key,
		"promotedTo": *to,
		"status":    "promoted",
	})
}

func cmdStats(args []string) {
	fs := flag.NewFlagSet("stats", flag.ExitOnError)
	dbPath := fs.String("db", "", "database path")
	fs.Parse(args)

	db, err := openDB(*dbPath)
	if err != nil {
		fatal("open db: %v", err)
	}
	defer db.Close()

	stats, err := db.Stats()
	if err != nil {
		fatal("stats failed: %v", err)
	}

	jsonOut(stats)
}
