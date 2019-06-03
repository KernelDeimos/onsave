package watcher

import (
	"errors"
	"os"
	"sync"

	"github.com/KernelDeimos/anything-gos/interp_a"
	"github.com/KernelDeimos/onsave/inputfilters"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
)

const (
	LogFormatFSN = "[%s] %s: %s"
	LogTag       = "server"
)

type WatcherExit struct{}

type TrackedFile struct {
	Interp interp_a.HybridEvaluator
	Script []interface{}
}

func (tf TrackedFile) Exec() {
	result, err := tf.Interp.OpEvaluate(tf.Script)
	log.Info(result)
	if err != nil {
		log.Error("[script] ", err)
	}
}

type Watcher struct {
	Execer  *interp_a.HybridEvaluator
	watcher *fsnotify.Watcher
	watched map[string]TrackedFile

	lockWatched *sync.RWMutex

	exit chan WatcherExit

	PathFilters inputfilters.PathFiltersI
}

func NewDefault() (*Watcher, error) {
	s := &Watcher{}

	var err error
	s.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return s, err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return s, errors.New("could not get cwd: " + err.Error())
	}
	s.PathFilters = inputfilters.PathFilters{
		WorkingDirectory: cwd,
	}

	s.watched = map[string]TrackedFile{}

	s.lockWatched = &sync.RWMutex{}

	return s, nil
}

func (s Watcher) Close() {
	s.watcher.Close()
}

func (s Watcher) Run() {

	for {
		select {
		case event := <-s.watcher.Events:
			// fsnotify gives some events without filenames;
			// ignore these.
			if event.Name == "" {
				continue
			}
			log.Infof(LogFormatFSN, LogTag, event.Name, event.String())
			entry, ok := s.getEntry(event.Name)
			if !ok {
				log.Errorf(
					"[%s] could not match event to watched list: event %#x on '%s'",
					"[watcher]",
					event.Op, event.Name)
			}

			if (event.Op & fsnotify.Write) != 0 {
				entry.Exec()
			}
			if (event.Op & fsnotify.Rename) != 0 {
				entry.Exec()
				s.watcher.Remove(event.Name)
				err := s.watcher.Add(event.Name)
				if err != nil {
					log.Warn(err)
				}

			}
			log.Info("Return")
		case err := <-s.watcher.Errors:
			log.Error("[fsnotify]", err)
		case <-s.exit:
			log.Warn("Exiting")
			return
		}
	}
}

func (s *Watcher) unwatch(npath string) {
	s.lockWatched.Lock()
	defer s.lockWatched.Unlock()

	delete(s.watched, npath)
}

func (s *Watcher) getEntry(npath string) (TrackedFile, bool) {
	s.lockWatched.RLock()
	defer s.lockWatched.RUnlock()

	a, b := s.watched[npath]
	return a, b
}

func (s *Watcher) setEntry(npath string, entry TrackedFile) {
	s.lockWatched.Lock()
	defer s.lockWatched.Unlock()

	s.watched[npath] = entry
}

func (s *Watcher) AddRule(
	ii interp_a.HybridEvaluator,
	path string, script []interface{},
) error {
	// Normalise path
	npath, err := s.PathFilters.NormaliseToRel(path)
	if err != nil {
		return err
	}

	path = "" // Using the user-provided path after this point is an error

	log.Debugf("preparing to track file '%s'=='%s'", path, npath)

	trackedFile := TrackedFile{
		Interp: ii,
		Script: script,
	}

	s.setEntry(npath, trackedFile)

	log.Info(npath)

	err = s.watcher.Add(npath)
	if err != nil {
		s.unwatch(npath)
		return err
	}

	return nil
}
