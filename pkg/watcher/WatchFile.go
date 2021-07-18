package watcher

import (
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
)

// Resilliant file watcher that handles file renames
// Special features:
// 1. This debounces multiple quick changes before invoking the callback
// 2. After the callback, resubscribe to the file to handle file renames that change the file inode
//
//  path to watch
//  handler to invoke on change
// This returns the fsnotify watcher. Close it when done.
func WatchFile(path string, handler func() error) (*fsnotify.Watcher, error) {
	watcher, _ := fsnotify.NewWatcher()
	// The callback timer debounces multiple changes to the config file
	callbackTimer := time.AfterFunc(0, func() {
		logrus.Debug("AuthStoreFile.Watch: invoking callback")
		handler()
		//
		// file renames change the inode of the filename, resubscribe
		watcher.Remove(path)
		watcher.Add(path)
	})
	callbackTimer.Stop() // don't start yet

	err := watcher.Add(path)
	if err != nil {
		logrus.Errorf("AuthStoreFile.Watch: unable to watch for changes: %s", err)
		return watcher, err
	}
	// defer watcher.Close()

	// done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// don't really care what the change it, 100msec after the last event the file will reload
				logrus.Debugf("Watch: event: %s. Modified file: %s", event, event.Name)
				callbackTimer.Reset(time.Millisecond * 100)
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logrus.Errorf("Watch: Error: %s", err)
			}
		}
	}()
	// err = watcher.Add(path)
	// if err != nil {
	// 	logrus.Errorf("Watch: error %s", err)
	// }
	// <-done
	return watcher, nil
}
