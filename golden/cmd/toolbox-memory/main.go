package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"toolbox-memory/internal/memory"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage(os.Stderr)
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
	case "peek":
		cmdPeek(os.Args[2:])
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
	case "namespaces":
		cmdNamespaces(os.Args[2:])
	case "version":
		fmt.Println(version)
	case "help", "--help", "-h":
		if len(os.Args) >= 3 {
			printHelpTopic(os.Args[2])
		} else {
			printUsage(os.Stdout)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", subcmd)
		printUsage(os.Stderr)
		os.Exit(1)
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, `Usage: toolbox-memory <command> [flags]

Commands:
  init      Create the database file
  write     Store a memory entry
  search    Full-text search within a namespace
  read      Get a single entry by namespace+key (increments hit_count)
  peek      Get a single entry by namespace+key (no side effects — read-only)
  list      List active/validated entries in a namespace
  delete    Delete an entry by namespace+key
  prune     Run lifecycle transitions (active->stale->archived)
  promote   Mark an entry as promoted
  stats     Print lifecycle statistics (--by-ns adds per-namespace breakdown)
  namespaces List all namespaces with per-lifecycle counts
  version   Print version

Global flag (all commands except version):
  --db PATH   Override database path (default: .claude/memory/toolbox.db)

Run 'toolbox-memory help <topic>' for conceptual topics.`)
	fmt.Fprintf(w, "Available topics: %s\n", strings.Join(helpTopics, ", "))
}

var helpTopics = []string{"lifecycle", "agent", "namespaces", "upsert", "examples"}

func printHelpTopic(topic string) {
	switch topic {
	case "lifecycle":
		fmt.Println(helpLifecycle)
	case "agent":
		fmt.Println(helpAgent)
	case "namespaces":
		fmt.Println(helpNamespaces)
	case "upsert":
		fmt.Println(helpUpsert)
	case "examples":
		fmt.Println(helpExamples)
	default:
		fmt.Fprintf(os.Stderr, "unknown help topic: %q\n\nAvailable topics: %s\n",
			topic, strings.Join(helpTopics, ", "))
		os.Exit(1)
	}
}

const helpLifecycle = `Lifecycle

Every entry moves through these states:

  active     — freshly written; hit_count = 0; confidence = 0.5
  validated  — proven by recurrence: active entries with hit_count >= 2
               transition here on the next 'prune' run
  promoted   — explicitly marked important via 'promote'; confidence = 1.0;
               the promoted_to column records the target (e.g. CLAUDE.md section)
  stale      — active with no hits in 60+ days; transitioned by 'prune'
  archived   — stale for 30+ more days; transitioned by 'prune'

Which states 'list' and 'search' return: active and validated only.
Which states 'stats' counts: all five, plus a total.
When to run 'prune': on /slim and /pomo invocations, or whenever stats shows
growth in the active bucket. It's idempotent and cheap for typical DBs.

Confidence is NOT mutated on read — it's 0.5 on write and 1.0 on promote.
Past versions auto-bumped confidence on every read/search; that behavior
was removed.`

const helpAgent = `Agent field

The 'agent' column is audit metadata: which agent emitted this entry.
It is NOT an access-control dimension and NOT a search filter — every
subcommand ignores it when matching (ns, key). The idiomatic values used
by the golden set agents:

  debugger   — writes bug_pattern
  planner    — writes spec_gap, calibration
  reviewer   — writes spec_gap
  builder    — writes calibration
  close-issue, pomo — audit tagging only

Upsert overwrite: 'write --ns X --key Y --agent Z' on an existing (X, Y)
silently overwrites the agent column with Z, along with the value. This is
intentional (the newest writer is authoritative for audit), but worth
knowing — if two agents are writing to the same key, the order of writes
determines who gets credit. See 'help upsert' for the full semantics.`

const helpNamespaces = `Namespaces

Valid namespaces in the golden set's toolbox-memory usage:

  bug_pattern  — debugger's record of bug classes + prevention strategies
  spec_gap     — what the spec should have included (reviewer, planner)
  calibration  — estimated vs actual for features (builder, planner)
  routing      — cross-agent orchestration signals (reserved)

These are agent-emitted signal. toolbox-memory is NOT for user-facing
lessons — those live in .claude/lessons.md (with an in-file markdown
lifecycle managed by /pomo) and in flat-file auto-memory (which the system
prompt handles). The three-way split is described in the repo README.

Discovery: 'toolbox-memory namespaces' lists every namespace actually
present in the DB, with per-lifecycle counts. 'toolbox-memory stats --by-ns'
combines lifecycle totals with per-namespace breakdown.`

const helpUpsert = `Upsert semantics

'write --ns X --key Y --agent Z --value V' is an upsert on (X, Y):

  - If no row exists: inserts a fresh row with confidence=0.5, hit_count=0,
    lifecycle=active.
  - If a row exists: replaces value AND agent; updates updated_at;
    PRESERVES hit_count and confidence.

The preserved hit_count is significant: 'prune' promotes active entries
to validated once hit_count >= 2, so rewriting a well-used entry does NOT
reset its progress toward validation.

If you want to track distinct incidents separately (instead of merging into
one row with a growing hit_count), include a discriminator in the key:

  --key "nil-pointer-2026-04-23"
  --key "nil-pointer-handler-foo"

Delete + rewrite is NOT a no-op — delete clears hit_count and confidence;
the rewrite starts fresh.`

const helpExamples = `Examples

Write with inline value:
  toolbox-memory write --ns bug_pattern --agent debugger \
    --key "nil-pointer-reset" --value '{"class":"state-corruption"}'

Write from a file (preserves bytes exactly):
  toolbox-memory write --ns spec_gap --agent reviewer \
    --key "login-flow-gap" --value-file ./gap.json

Write from stdin:
  echo -n '{"rule":"x"}' | toolbox-memory write --ns bug_pattern \
    --agent debugger --key inline-rule --value -

Search (hyphenated tokens auto-quoted):
  toolbox-memory search --ns bug_pattern --query "alt-screen logger"

Search with native FTS5 operators:
  toolbox-memory search --ns bug_pattern --query "alpha -beta" --raw

Peek (side-effect-free read — no hit_count bump):
  toolbox-memory peek --ns bug_pattern --key nil-pointer-reset

Promote (value column untouched; promoted_to set):
  toolbox-memory promote --ns bug_pattern --key nil-pointer-reset \
    --to "CLAUDE.md Error Handling"

Stats with per-namespace breakdown:
  toolbox-memory stats --by-ns

List every namespace in the DB:
  toolbox-memory namespaces

Run lifecycle transitions (active->validated, active->stale, stale->archived):
  toolbox-memory prune`

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

// withExample overrides fs.Usage to include an example block after the
// auto-generated flag listing. Agents and humans running 'toolbox-memory
// <cmd> --help' see a concrete invocation alongside the flag syntax.
func withExample(fs *flag.FlagSet, example string) {
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", fs.Name())
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n  %s\n", example)
	}
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
	withExample(fs, "toolbox-memory init --db /tmp/toolbox.db")
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
	value := fs.String("value", "", "value (use '-' for stdin; mutually exclusive with --value-file)")
	valueFile := fs.String("value-file", "", "read value from PATH (mutually exclusive with --value)")
	withExample(fs, `toolbox-memory write --ns bug_pattern --agent debugger --key nil-pointer-reset --value '{"class":"state-corruption"}'`)
	fs.Parse(args)

	if *ns == "" || *agent == "" || *key == "" {
		fatal("write requires --ns, --agent, and --key")
	}

	valueProvided := *value != "" && *value != "-"
	stdinRequested := *value == "-"
	fileProvided := *valueFile != ""

	set := 0
	if valueProvided {
		set++
	}
	if stdinRequested {
		set++
	}
	if fileProvided {
		set++
	}
	if set != 1 {
		fatal("write requires exactly one of --value STRING, --value -, or --value-file PATH")
	}

	resolved := *value
	switch {
	case stdinRequested:
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			fatal("read stdin: %v", err)
		}
		resolved = string(b)
	case fileProvided:
		b, err := os.ReadFile(*valueFile)
		if err != nil {
			fatal("read --value-file: %v", err)
		}
		resolved = string(b)
	}

	db, err := openDB(*dbPath)
	if err != nil {
		fatal("open db: %v", err)
	}
	defer db.Close()

	id, err := db.Store(*ns, *agent, *key, resolved)
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
	raw := fs.Bool("raw", false, "pass query to FTS5 unchanged (bypass hyphen-safe sanitization)")
	withExample(fs, `toolbox-memory search --ns bug_pattern --query "alt-screen logger"`)
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

	entries, err := db.Search(*ns, *query, limit, *raw)
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
	withExample(fs, "toolbox-memory read --ns bug_pattern --key nil-pointer-reset")
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

func cmdPeek(args []string) {
	fs := flag.NewFlagSet("peek", flag.ExitOnError)
	dbPath := fs.String("db", "", "database path")
	ns := fs.String("ns", "", "namespace (required)")
	key := fs.String("key", "", "key (required)")
	withExample(fs, "toolbox-memory peek --ns bug_pattern --key nil-pointer-reset")
	fs.Parse(args)

	if *ns == "" || *key == "" {
		fatal("peek requires --ns and --key")
	}

	db, err := openDB(*dbPath)
	if err != nil {
		fatal("open db: %v", err)
	}
	defer db.Close()

	entry, err := db.Peek(*ns, *key)
	if err != nil {
		fatal("peek failed: %v", err)
	}

	jsonOut(entry)
}

func cmdList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	dbPath := fs.String("db", "", "database path")
	ns := fs.String("ns", "", "namespace (required)")
	withExample(fs, "toolbox-memory list --ns bug_pattern")
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
	withExample(fs, "toolbox-memory delete --ns bug_pattern --key stale-pattern")
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
	withExample(fs, "toolbox-memory prune")
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
	withExample(fs, `toolbox-memory promote --ns bug_pattern --key nil-pointer-reset --to "CLAUDE.md Error Handling"`)
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
	byNs := fs.Bool("by-ns", false, "include per-namespace breakdown")
	withExample(fs, "toolbox-memory stats --by-ns")
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

	if !*byNs {
		jsonOut(stats)
		return
	}

	byNamespace, err := db.Namespaces()
	if err != nil {
		fatal("stats --by-ns: %v", err)
	}
	jsonOut(map[string]interface{}{
		"lifecycle":   stats,
		"byNamespace": byNamespace,
	})
}

func cmdNamespaces(args []string) {
	fs := flag.NewFlagSet("namespaces", flag.ExitOnError)
	dbPath := fs.String("db", "", "database path")
	withExample(fs, "toolbox-memory namespaces")
	fs.Parse(args)

	db, err := openDB(*dbPath)
	if err != nil {
		fatal("open db: %v", err)
	}
	defer db.Close()

	nss, err := db.Namespaces()
	if err != nil {
		fatal("namespaces failed: %v", err)
	}
	jsonOut(nss)
}
