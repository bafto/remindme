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

	"github.com/google/uuid"
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
	Id    uuid.UUID
}

func NewEntry(when time.Time, title, msg string) Entry {
	return Entry{
		When:  when.Format(timeLayout),
		Title: title,
		Msg:   msg,
		Id:    uuid.New(),
	}
}

func (e *Entry) GetTime() (time.Time, error) {
	return time.Parse(timeLayout, e.When)
}

func overrideReminders(entries []Entry) error {
	data, err := json.Marshal(entries)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(savePath, os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(data)
	return err
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

func AddReminder(entry Entry) error {
	if entries, err := GetAllReminders(); err != nil {
		return err
	} else {
		entries = append(entries, entry)
		return overrideReminders(entries)
	}
}

// removes the specified entry by comparing it
// returns the number of remaining reminders
func RemoveReminder(entry Entry) (int, error) {
	entries, err := GetAllReminders()
	if err != nil {
		return 0, err
	}

	// remove the specified entry by checking
	for i, other := range entries {
		if other.Id == entry.Id {
			entries = append(entries[:i], entries[i+1:]...)
		}
	}

	return len(entries), overrideReminders(entries)
}
