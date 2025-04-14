package analyzers

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/pkgfact"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"
)

/*
PassesAnalyzers returns set of standard go analyzers frm x/tools/go/analysis.
appends 			Package appends defines an Analyzer that detects if there is only one variable in append.
asmdecl 			Package asmdecl defines an Analyzer that reports mismatches between assembly files and Go declarations.
assign 				Package assign defines an Analyzer that detects useless assignments.
atomic 				Package atomic defines an Analyzer that checks for common mistakes using the sync/atomic package.
atomicalign 		Package atomicalign defines an Analyzer that checks for non-64-bit-aligned arguments to sync/atomic functions.
bools 				Package bools defines an Analyzer that detects common mistakes involving boolean operators.
buildssa 			Package buildssa defines an Analyzer that constructs the SSA representation of an error-free package and returns the set of all functions within it.
buildtag 			Package buildtag defines an Analyzer that checks build tags.
cgocall 			Package cgocall defines an Analyzer that detects some violations of the cgo pointer passing rules.
composite 			Package composite defines an Analyzer that checks for unkeyed composite literals.
copylock 			Package copylock defines an Analyzer that checks for locks erroneously passed by value.
ctrlflow 			Package ctrlflow is an analysis that provides a syntactic control-flow graph (CFG) for the body of a function.
deepequalerrors 	Package deepequalerrors defines an Analyzer that checks for the use of reflect.DeepEqual with error values.
defers 				Package defers defines an Analyzer that checks for common mistakes in defer statements.
directive 			Package directive defines an Analyzer that checks known Go toolchain directives.
errorsas 			The errorsas package defines an Analyzer that checks that the second argument to errors.As is a pointer to a type implementing error.
fieldalignment 		Package fieldalignment defines an Analyzer that detects structs that would use less memory if their fields were sorted.
findcall 			Package findcall defines an Analyzer that serves as a trivial example and test of the Analysis API.
framepointer 		Package framepointer defines an Analyzer that reports assembly code that clobbers the frame pointer before saving it.
httpresponse 		Package httpresponse defines an Analyzer that checks for mistakes using HTTP responses.
ifaceassert 		Package ifaceassert defines an Analyzer that flags impossible interface-interface type assertions.
inspect 			Package inspect defines an Analyzer that provides an AST inspector (golang.org/x/tools/go/ast/inspector.Inspector) for the syntax trees of a package.
loopclosure 		Package loopclosure defines an Analyzer that checks for references to enclosing loop variables from within nested functions.
lostcancel 			Package lostcancel defines an Analyzer that checks for failure to call a context cancellation function.
nilfunc 			Package nilfunc defines an Analyzer that checks for useless comparisons against nil.
nilness 			Package nilness inspects the control-flow graph of an SSA function and reports errors such as nil pointer dereferences and degenerate nil pointer comparisons.
pkgfact 			The pkgfact package is a demonstration and test of the package fact mechanism.
printf 				Package printf defines an Analyzer that checks consistency of Printf format strings and arguments.
reflectvaluecompare Package reflectvaluecompare defines an Analyzer that checks for accidentally using == or reflect.DeepEqual to compare reflect.Value values.
shadow 				Package shadow defines an Analyzer that checks for shadowed variables.
shift 				Package shift defines an Analyzer that checks for shifts that exceed the width of an integer.
sigchanyzer 		Package sigchanyzer defines an Analyzer that detects misuse of unbuffered signal as argument to signal.Notify.
slog 				Package slog defines an Analyzer that checks for mismatched key-value pairs in log/slog calls.
sortslice 			Package sortslice defines an Analyzer that checks for calls to sort.Slice that do not use a slice type as first argument.
stdmethods 			Package stdmethods defines an Analyzer that checks for misspellings in the signatures of methods similar to well-known interfaces.
stdversion 			Package stdversion reports uses of standard library symbols that are "too new" for the Go version in force in the referring file.
stringintconv 		Package stringintconv defines an Analyzer that flags type conversions from integers to strings.
structtag 			Package structtag defines an Analyzer that checks struct field tags are well formed.
testinggoroutine 	Package testinggoroutine defines an Analyzerfor detecting calls to Fatal from a test goroutine.
tests 				Package tests defines an Analyzer that checks for common mistaken usages of tests and examples.
timeformat 			Package timeformat defines an Analyzer that checks for the use of time.Format or time.Parse calls with a bad format.
unmarshal 			The unmarshal package defines an Analyzer that checks for passing non-pointer or non-interface types to unmarshal and decode functions.
unreachable 		Package unreachable defines an Analyzer that checks for unreachable code.
unsafeptr 			Package unsafeptr defines an Analyzer that checks for invalid conversions of uintptr to unsafe.Pointer.
unusedresult 		Package unusedresult defines an analyzer that checks for unused results of calls to certain pure functions.
unusedwrite 		Package unusedwrite checks for unused writes to the elements of a struct or array object.
usesgenerics 		Package usesgenerics defines an Analyzer that checks for usage of generic features added in Go 1.18.
waitgroup 			Package waitgroup defines an Analyzer that detects simple misuses of sync.WaitGroup.
*/
func PassesAnalyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		atomicalign.Analyzer,
		bools.Analyzer,
		buildssa.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		ctrlflow.Analyzer,
		deepequalerrors.Analyzer,
		directive.Analyzer,
		errorsas.Analyzer,
		fieldalignment.Analyzer,
		findcall.Analyzer,
		framepointer.Analyzer,
		httpresponse.Analyzer,
		ifaceassert.Analyzer,
		inspect.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		nilness.Analyzer,
		pkgfact.Analyzer,
		printf.Analyzer,
		reflectvaluecompare.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		sigchanyzer.Analyzer,
		sortslice.Analyzer,
		stdmethods.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		testinggoroutine.Analyzer,
		tests.Analyzer,
		timeformat.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
		unusedwrite.Analyzer,
		usesgenerics.Analyzer,
	}
}
