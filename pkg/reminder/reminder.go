// package reminder handles the interaction
// with saved data
package reminder

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	timeLayout = time.UnixDate
)

var (
	savePath string
)

// make sure the save file exists
func init() {
	if path, err := os.UserCacheDir(); err == nil {
		savePath = filepath.Join(path, "/remindme/remindme.json")
	} else {
		log.Fatal(err)
	}

	if _, err := os.Stat(savePath); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(filepath.Dir(savePath), os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
		err = os.WriteFile(savePath, []byte("[]"), os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// represents a single entry in the saved reminders
type Entry struct {
	When  string
	Title string
	Msg   string
}

func (e *Entry) GetTime() (time.Time, error) {
	return time.Parse(timeLayout, e.When)
}

func GetAllReminders() (entries []Entry, err error) {
	f, err := os.OpenFile(savePath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&entries)
	if err != nil {
		return nil, err
	}
	return entries, nil
}
