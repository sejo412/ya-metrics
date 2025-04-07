package analyzers

import (
	"golang.org/x/tools/go/analysis"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

/*
StaticCheckSAAnalyzers returns all SA-grouped analyzers from `staticcheck` module.

SA1000	Invalid regular expression
SA1001	Invalid template
SA1002	Invalid format in time.Parse
SA1003	Unsupported argument to functions in encoding/binary
SA1004	Suspiciously small untyped constant in time.Sleep
SA1005	Invalid first argument to exec.Command
SA1006	Printf with dynamic first argument and no further arguments
SA1007	Invalid URL in net/url.Parse
SA1008	Non-canonical key in http.Header map
SA1010	(*regexp.Regexp).FindAll called with n == 0, which will always return zero results
SA1011	Various methods in the strings package expect valid UTF-8, but invalid input is provided
SA1012	A nil context.Context is being passed to a function, consider using context.TODO instead
SA1013	io.Seeker.Seek is being called with the whence constant as the first argument, but it should be the second
SA1014	Non-pointer value passed to Unmarshal or Decode
SA1015	Using time.Tick in a way that will leak. Consider using time.NewTicker, and only use time.Tick in tests, commands and endless functions
SA1016	Trapping a signal that cannot be trapped
SA1017	Channels used with os/signal.Notify should be buffered
SA1018	strings.Replace called with n == 0, which does nothing
SA1019	Using a deprecated function, variable, constant or field
SA1020	Using an invalid host:port pair with a net.Listen-related function
SA1021	Using bytes.Equal to compare two net.IP
SA1023	Modifying the buffer in an io.Writer implementation
SA1024	A string cutset contains duplicate characters
SA1025	It is not possible to use (*time.Timer).Reset’s return value correctly
SA1026	Cannot marshal channels or functions
SA1027	Atomic access to 64-bit variable must be 64-bit aligned
SA1028	sort.Slice can only be used on slices
SA1029	Inappropriate key in call to context.WithValue
SA1030	Invalid argument in call to a strconv function
SA1031	Overlapping byte slices passed to an encoder
SA1032	Wrong order of arguments to errors.Is
SA2000	sync.WaitGroup.Add called inside the goroutine, leading to a race condition
SA2001	Empty critical section, did you mean to defer the unlock?
SA2002	Called testing.T.FailNow or SkipNow in a goroutine, which isn’t allowed
SA2003	Deferred Lock right after locking, likely meant to defer Unlock instead
SA3000	TestMain doesn’t call os.Exit, hiding test failures
SA3001	Assigning to b.N in benchmarks distorts the results
SA4000	Binary operator has identical expressions on both sides
SA4001	&*x gets simplified to x, it does not copy x
SA4003	Comparing unsigned values against negative values is pointless
SA4004	The loop exits unconditionally after one iteration
SA4005	Field assignment that will never be observed. Did you mean to use a pointer receiver?
SA4006	A value assigned to a variable is never read before being overwritten. Forgotten error check or dead code?
SA4008	The variable in the loop condition never changes, are you incrementing the wrong variable?
SA4009	A function argument is overwritten before its first use
SA4010	The result of append will never be observed anywhere
SA4011	Break statement with no effect. Did you mean to break out of an outer loop?
SA4012	Comparing a value against NaN even though no value is equal to NaN
SA4013	Negating a boolean twice (!!b) is the same as writing b. This is either redundant, or a typo.
SA4014	An if/else if chain has repeated conditions and no side-effects; if the condition didn’t match the first time, it won’t match the second time, either
SA4015	Calling functions like math.Ceil on floats converted from integers doesn’t do anything useful
SA4016	Certain bitwise operations, such as x ^ 0, do not do anything useful
SA4017	Discarding the return values of a function without side effects, making the call pointless
SA4018	Self-assignment of variables
SA4019	Multiple, identical build constraints in the same file
SA4020	Unreachable case clause in a type switch
SA4021	x = append(y) is equivalent to x = y
SA4022	Comparing the address of a variable against nil
SA4023	Impossible comparison of interface value with untyped nil
SA4024	Checking for impossible return value from a builtin function
SA4025	Integer division of literals that results in zero
SA4026	Go constants cannot express negative zero
SA4027	(*net/url.URL).Query returns a copy, modifying it doesn’t change the URL
SA4028	x % 1 is always zero
SA4029	Ineffective attempt at sorting slice
SA4030	Ineffective attempt at generating random number
SA4031	Checking never-nil value against nil
SA4032	Comparing runtime.GOOS or runtime.GOARCH against impossible value
SA5000	Assignment to nil map
SA5001	Deferring Close before checking for a possible error
SA5002	The empty for loop (for {}) spins and can block the scheduler
SA5003	Defers in infinite loops will never execute
SA5004	for { select { ... with an empty default branch spins
SA5005	The finalizer references the finalized object, preventing garbage collection
SA5007	Infinite recursive call
SA5008	Invalid struct tag
SA5009	Invalid Printf call
SA5010	Impossible type assertion
SA5011	Possible nil pointer dereference
SA5012	Passing odd-sized slice to function expecting even size
SA6000	Using regexp.Match or related in a loop, should use regexp.Compile
SA6001	Missing an optimization opportunity when indexing maps by byte slices
SA6002	Storing non-pointer values in sync.Pool allocates memory
SA6003	Converting a string to a slice of runes before ranging over it
SA6005	Inefficient string comparison with strings.ToLower or strings.ToUpper
SA6006	Using io.WriteString to write []byte
SA9001	Defers in range loops may not run when you expect them to
SA9002	Using a non-octal os.FileMode that looks like it was meant to be in octal.
SA9003	Empty body in an if or else branch
SA9004	Only the first constant has an explicit type
SA9005	Trying to marshal a struct with no public fields nor custom marshaling
SA9006	Dubious bit shifting of a fixed size integer value
SA9007	Deleting a directory that shouldn’t be deleted
SA9008	else branch of a type assertion is probably not reading the right value
SA9009	Ineffectual Go compiler directive
S1011	Use a single append to concatenate two slices
*/
func StaticCheckSAAnalyzers() []*analysis.Analyzer {
	result := make([]*analysis.Analyzer, 0, len(staticcheck.Analyzers))
	for _, v := range staticcheck.Analyzers {
		result = append(result, v.Analyzer)
	}
	return result
}

// StaticCheckSAnalyzer returns S1011 analyzer from `staticcheck` module.
// S1011	Use a single append to concatenate two slices
func StaticCheckSAnalyzer() *analysis.Analyzer {
	for _, v := range simple.Analyzers {
		if v.Analyzer.Name == "S1011" {
			return v.Analyzer
		}
	}
	return &analysis.Analyzer{}
}

// StaticCheckSTAnalyzer returns ST1005 analyzer from `staticcheck` module.
// ST1005	Incorrectly formatted error string
func StaticCheckSTAnalyzer() *analysis.Analyzer {
	for _, v := range stylecheck.Analyzers {
		if v.Analyzer.Name == "ST1005" {
			return v.Analyzer
		}
	}
	return &analysis.Analyzer{}
}

// StaticCheckQFAnalyzer returns QF1012 analyzer from `staticcheck` module.
// QF1012	Use fmt.Fprintf(x, ...) instead of x.Write(fmt.Sprintf(...))
func StaticCheckQFAnalyzer() *analysis.Analyzer {
	for _, v := range quickfix.Analyzers {
		if v.Analyzer.Name == "QF1012" {
			return v.Analyzer
		}
	}
	return &analysis.Analyzer{}
}
