package main

import "testing"

func TestGenFilename(t *testing.T) {
	type genFilenameTest struct {
		In1 string
		In2 string
		Want string
	}

	tests := []genFilenameTest{
		{"test", ".ext", "test.ext"},
		{"test.other", ".ext", "test.ext"},
		{"important_dialog.txt", ".dialog", "important_dialog.dialog"},
	}

	for _, test := range tests {
		if output := genFilename(test.In1, test.In2); output != test.Want {
			t.Errorf("Output %q not equal to %q", output, test.Want)
		}
	}
}

func TestValidateLines(t *testing.T) {
	type validateLinesTest struct {
		In1 string
		ShouldSucceed bool
	}

	tests := []validateLinesTest{
		{`String string string string string string
String string string string string string
String string string string string string
String string string string string string`, true},
		{`String string string string string string string`, false},
		{`String string string string string string
String string string string string string
String string string string string string string
String string string string string string`, false},
	}

	for _, test := range tests {
		output := validateLines(test.In1)
		if (test.ShouldSucceed && output != nil) || (!test.ShouldSucceed && output == nil) {
			negative := " "
			if !test.ShouldSucceed {
				negative = " not "
			}
			t.Errorf("Input %q should" + negative + "succeed, but it did not", test.In1)
		}
	}
}
