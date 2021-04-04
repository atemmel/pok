package pok

import(
	"github.com/sqweek/dialog"
	"os"
	"runtime/debug"
)

var doLogToFile bool
var dieOnAssert bool
var logFileName string

func InitAssert(path *string, die bool) {
	if path != nil {
		doLogToFile = true
		logFileName = *path
		os.Remove(*path)
	} else {
		doLogToFile = false
	}

	dieOnAssert = die
}

func Assert(condition error) {
	if condition == nil {
		return
	}

	msg := make([]byte, 0, 256)
	msg = append(msg, []byte(condition.Error())...)
	msg = append(msg, '\n')
	msg = append(msg, debug.Stack()...)
	msg = append(msg, '\n')

	if doLogToFile {
		file, err := os.OpenFile(logFileName, os.O_APPEND | os.O_WRONLY | os.O_CREATE, 0644)
		if err != nil {
			dialog.Message("Could not open file to log error!").Title("Game error").Error()
		}
		defer file.Close()

		if err == nil {
			file.Write(msg)
		}
	}

	dialog.Message("Assertion reached: %s", condition.Error()).Title("Game error").Error()

	if dieOnAssert {
		panic(condition)
	}
}
