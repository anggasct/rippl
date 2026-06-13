package testmap

import "strings"

func isTestFile(path string) bool {
	return strings.HasSuffix(path, "_test.go")
}

func sourcePathForTest(testPath string) (string, bool) {
	if !isTestFile(testPath) {
		return "", false
	}
	return testPath[:len(testPath)-len("_test.go")] + ".go", true
}

func appendTestFile(out map[string]FileCoverage, sourcePath, testPath string) {
	fc := out[sourcePath]
	fc.Path = sourcePath
	fc.HasTestFile = true
	fc.TestFiles = appendUnique(fc.TestFiles, testPath)
	if fc.Status == StatusNoTest {
		fc.Status = StatusUnknown
	}
	out[sourcePath] = fc
}

func appendUnique(list []string, item string) []string {
	for _, v := range list {
		if v == item {
			return list
		}
	}
	return append(list, item)
}
