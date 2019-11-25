package testcases

var AvailableTests = make(map[string]Tester)

func register(t Tester) {
	AvailableTests[t.Name()] = t
}
