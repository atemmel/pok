package pok

import(
	"io/ioutil"
	"encoding/json"
)

//TODO: Improve this, dialog tree maybe?
// https://i.stack.imgur.com/fkNWM.png
type Dialog struct {
	Dialog string
}

func ReadDialogFromFile(path string) Dialog {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return Dialog{
			"Could not load: '" + path + "'",
		}
	}

	var dialog Dialog
	err = json.Unmarshal(data, &dialog)
	if err != nil {
		return Dialog{
			"Could not parse: '" + path + "'",
		}
	}

	return dialog
}
