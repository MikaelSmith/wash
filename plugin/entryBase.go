package plugin

import (
	"context"
	"flag"
	"strings"
	"time"
)

type actionOpCode int8

const (
	// List represents Group#List
	List actionOpCode = iota
	// Open represents Readable#Open
	Open
	// Metadata represents Resource#Metadata
	Metadata
)

var actionOpCodeToNameMap = [3]string{"List", "Open", "Metadata"}

// EntryBase implements Entry, making it easy to create new entries.
// You should use plugin.NewEntry to create new EntryBase objects.
type EntryBase struct {
	name               string
	canonicalName      string
	slashReplacementCh rune
	// washID represents the entry's wash ID. It is set in CachedList.
	washID string
	ttl    [3]time.Duration
}

// newEntryBase is needed by NewEntry, NewRegistry,
// and some of the cache tests
func newEntryBase(name string, slashReplacementCh rune) EntryBase {
	e := EntryBase{
		name:               name,
		slashReplacementCh: slashReplacementCh,
		canonicalName:      strings.Replace(name, "/", string(slashReplacementCh), -1),
	}

	for op := range e.ttl {
		e.SetTTLOf(actionOpCode(op), 15*time.Second)
	}

	return e
}

// NewEntry creates a new entry, and replaces any '/' with '#'
func NewEntry(name string) EntryBase {
	return NewEntryWithSlashReplacementChar(name, '#')
}

// NewEntryWithSlashReplacementChar creates a new entry, and replaces any '/' with the specified rune
func NewEntryWithSlashReplacementChar(name string, slashReplacementCh rune) EntryBase {
	if name == "" {
		panic("plugin.NewEntry: received an empty name")
	}

	return newEntryBase(name, slashReplacementCh)
}

// ENTRY INTERFACE

// Name returns the entry's name as it was passed into
// plugin.NewEntry. You should use e.Name() when making
// the appropriate API calls within your plugin.
func (e *EntryBase) Name() string {
	return e.name
}

func (e *EntryBase) cname() string {
	return e.canonicalName
}

// Attr returns the entry's attributes. The default return value
// is a zero'ed attributes struct with the Size field set to
// SizeUnknown. You should override Attr() if you'd like to return
// a different set of attributes.
//
// NOTE: See the comments for plugin.Attr if you're interested in
// knowing why the Size field is set to SizeUnknown.
func (e *EntryBase) Attr(ctx context.Context) (Attributes, error) {
	return Attributes{
		Size: SizeUnknown,
	}, nil
}

func (e *EntryBase) slashReplacementChar() rune {
	return e.slashReplacementCh
}

func (e *EntryBase) id() string {
	return e.washID
}

func (e *EntryBase) setID(id string) {
	e.washID = id
}

func (e *EntryBase) getTTLOf(op actionOpCode) time.Duration {
	return e.ttl[op]
}

// OTHER METHODS USED TO FACILITATE PLUGIN DEVELOPMENT
// AND TESTING

// SetTTLOf sets the specified op's TTL
func (e *EntryBase) SetTTLOf(op actionOpCode, ttl time.Duration) {
	e.ttl[op] = ttl
}

// DisableCachingFor disables caching for the specified op
func (e *EntryBase) DisableCachingFor(op actionOpCode) {
	e.SetTTLOf(op, -1)
}

// DisableDefaultCaching disables the default caching
// for List, Open and Metadata.
func (e *EntryBase) DisableDefaultCaching() {
	for op := range e.ttl {
		e.DisableCachingFor(actionOpCode(op))
	}
}

// SetTestID sets the entry's cache ID for testing.
// It can only be called by the tests.
func (e *EntryBase) SetTestID(id string) {
	if notRunningTests() {
		panic("SetTestID can be only be called by the tests")
	}

	e.setID(id)
}

func notRunningTests() bool {
	return flag.Lookup("test.v") == nil
}
