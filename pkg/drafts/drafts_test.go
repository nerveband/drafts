package drafts

import (
	"fmt"
	"testing"
	"time"

	"github.com/ernstwi/drafts/internal/assert"
)

func TestCreateDefault(t *testing.T) {
	text := rand()
	uuid := Create(text, CreateOptions{})
	defer func() {
		Trash(uuid)
	}()
	draft := Get(uuid)
	assert.Equal(t, uuid, draft.UUID)
	assert.Equal(t, text, draft.Content)
	assert.EqualSlice(t, []string{}, draft.Tags)
	assert.Equal(t, false, draft.IsFlagged)
	assert.Equal(t, false, draft.IsArchived)
	assert.Equal(t, false, draft.IsTrashed)
}

func TestCreateFlagged(t *testing.T) {
	text := rand()
	uuid := Create(text, CreateOptions{Flagged: true})
	defer func() {
		Trash(uuid)
	}()
	draft := Get(uuid)
	assert.Equal(t, uuid, draft.UUID)
	assert.Equal(t, text, draft.Content)
	assert.EqualSlice(t, []string{}, draft.Tags)
	assert.Equal(t, true, draft.IsFlagged)
}

func TestCreateArchived(t *testing.T) {
	text := rand()
	uuid := Create(text, CreateOptions{Folder: FolderArchive})
	defer func() {
		Trash(uuid)
	}()
	draft := Get(uuid)
	assert.Equal(t, uuid, draft.UUID)
	assert.Equal(t, text, draft.Content)
	assert.EqualSlice(t, []string{}, draft.Tags)
	assert.Equal(t, true, draft.IsArchived)
}

func TestCreateTags(t *testing.T) {
	text := rand()
	tag := rand()
	uuid := Create(text, CreateOptions{Tags: []string{tag}})
	defer func() {
		Trash(uuid)
	}()
	draft := Get(uuid)
	assert.Equal(t, uuid, draft.UUID)
	assert.Equal(t, text, draft.Content)
	assert.EqualSlice(t, []string{tag}, draft.Tags)
}

func TestPrepend(t *testing.T) {
	text := rand()
	prefix := rand()
	uuid := Create(text, CreateOptions{})
	defer func() {
		Trash(uuid)
	}()
	Prepend(uuid, prefix, ModifyOptions{})
	content := Get(uuid).Content
	assert.Equal(t, prefix+"\n"+text, content)
}

func TestAppend(t *testing.T) {
	text := rand()
	suffix := rand()
	uuid := Create(text, CreateOptions{})
	defer func() {
		Trash(uuid)

	}()
	Append(uuid, suffix, ModifyOptions{})
	content := Get(uuid).Content
	assert.Equal(t, text+"\n"+suffix, content)
}

func TestReplace(t *testing.T) {
	text := rand()
	replacement := rand()
	uuid := Create(text, CreateOptions{})
	defer func() {
		Trash(uuid)
	}()
	Replace(uuid, replacement)
	content := Get(uuid).Content
	assert.Equal(t, replacement, content)
}

func TestTrash(t *testing.T) {
	text := rand()
	uuid := Create(text, CreateOptions{})
	Trash(uuid)
	draft := Get(uuid)
	assert.Equal(t, uuid, draft.UUID)
	assert.Equal(t, text, draft.Content)
	assert.EqualSlice(t, []string{}, draft.Tags)
	assert.Equal(t, true, draft.IsTrashed)
}

func TestArchive(t *testing.T) {
	text := rand()
	uuid := Create(text, CreateOptions{})
	Archive(uuid)
	draft := Get(uuid)
	assert.Equal(t, uuid, draft.UUID)
	assert.Equal(t, text, draft.Content)
	assert.EqualSlice(t, []string{}, draft.Tags)
	assert.Equal(t, true, draft.IsArchived)
}

func TestQuery(t *testing.T) {
	a := Create("A", CreateOptions{Tags: []string{"test", "a"}})
	b := Create("B", CreateOptions{Tags: []string{"test", "b"}, Flagged: true})
	c := Create("C", CreateOptions{Tags: []string{"test", "c"}})

	defer func() {
		for _, uuid := range []string{a, b, c} {
			Trash(uuid)
		}
	}()

	uuids := getUUIDs(Query("", FilterInbox, QueryOptions{Tags: []string{"test"}}))
	assert.EqualSlice(t, []string{a, b, c}, uuids)

	uuids = getUUIDs(Query("", FilterInbox, QueryOptions{Tags: []string{"test", "a"}}))
	assert.EqualSlice(t, []string{a}, uuids)

	uuids = getUUIDs(Query("", FilterInbox, QueryOptions{
		Tags:     []string{"test"},
		OmitTags: []string{"a"},
	}))
	assert.EqualSlice(t, []string{b, c}, uuids)

	// TODO: Testing Sort requires draft modification

	uuids = getUUIDs(Query("", FilterInbox, QueryOptions{
		Tags:           []string{"test"},
		SortDescending: true,
	}))
	assert.EqualSlice(t, []string{c, b, a}, uuids)

	uuids = getUUIDs(Query("", FilterInbox, QueryOptions{
		Tags:             []string{"test"},
		SortFlaggedToTop: true,
	}))
	assert.EqualSlice(t, []string{b, a, c}, uuids)
}

func TestSelect(t *testing.T) {
	a := Create("a", CreateOptions{})
	b := Create("b", CreateOptions{})
	defer func() {
		Trash(a)
		Trash(b)
	}()
	b_ := Get(Active()).Content
	Select(a)
	a_ := Get(Active()).Content
	assert.Equal(t, "a", a_)
	assert.Equal(t, "b", b_)
}

func TestGetSpecialChars(t *testing.T) {
	t.Skip()
	// https://en.wikipedia.org/wiki/URL_encoding#Percent-encoding_reserved_characters
	chars := []string{"␣", "!", "\"", "#", "$", "%", "&", "'", "(", ")", "*", "+", ",", "/", ":", ";", "=", "?", "@", "[", "]"}
	for _, c := range chars {
		uuid := Create(c, CreateOptions{})
		defer func() {
			Trash(uuid)
		}()
		content := Get(uuid).Content
		assert.Equal(t, c, content)
	}
}

func TestTag(t *testing.T) {
	text := rand()
	tag := rand()
	uuid := Create(text, CreateOptions{})
	defer func() {
		Trash(uuid)
	}()
	Tag(uuid, tag)
	draft := Get(uuid)
	assert.Equal(t, uuid, draft.UUID)
	assert.Equal(t, text, draft.Content)
	assert.EqualSlice(t, []string{tag}, draft.Tags)
}

func TestGetReturnsLanguageGrammar(t *testing.T) {
	text := rand()
	uuid := Create(text, CreateOptions{})
	defer func() {
		Trash(uuid)
	}()
	draft := Get(uuid)
	// languageGrammar is not exposed via AppleScript dictionary,
	// so the field will be empty when fetched via Get().
	// The struct field exists for JSON serialization compatibility.
	assert.Equal(t, "", draft.LanguageGrammar)
}

func TestGetReturnsLocationFields(t *testing.T) {
	text := rand()
	uuid := Create(text, CreateOptions{})
	defer func() {
		Trash(uuid)
	}()
	draft := Get(uuid)
	// Location fields are float64; verify they parsed successfully.
	// Values may be zero or actual coordinates depending on device location services.
	if draft.CreatedLatitude < -90 || draft.CreatedLatitude > 90 {
		t.Errorf("CreatedLatitude out of range: %f", draft.CreatedLatitude)
	}
	if draft.CreatedLongitude < -180 || draft.CreatedLongitude > 180 {
		t.Errorf("CreatedLongitude out of range: %f", draft.CreatedLongitude)
	}
	if draft.ModifiedLatitude < -90 || draft.ModifiedLatitude > 90 {
		t.Errorf("ModifiedLatitude out of range: %f", draft.ModifiedLatitude)
	}
	if draft.ModifiedLongitude < -180 || draft.ModifiedLongitude > 180 {
		t.Errorf("ModifiedLongitude out of range: %f", draft.ModifiedLongitude)
	}
}

func TestSetFlagged(t *testing.T) {
	text := rand()
	uuid := Create(text, CreateOptions{})
	defer func() {
		Trash(uuid)
	}()

	// Verify starts unflagged
	draft := Get(uuid)
	assert.Equal(t, false, draft.IsFlagged)

	// Flag it
	SetFlagged(uuid, true)
	draft = Get(uuid)
	assert.Equal(t, true, draft.IsFlagged)

	// Unflag it
	SetFlagged(uuid, false)
	draft = Get(uuid)
	assert.Equal(t, false, draft.IsFlagged)
}

func TestSetLanguageGrammar(t *testing.T) {
	text := rand()
	uuid := Create(text, CreateOptions{})
	defer func() {
		Trash(uuid)
	}()

	// SetLanguageGrammar should not panic.
	// languageGrammar is not readable via AppleScript, so we cannot
	// verify the round-trip. We confirm the function executes without error
	// and that the draft is otherwise unaffected.
	SetLanguageGrammar(uuid, "JavaScript")
	draft := Get(uuid)
	assert.Equal(t, uuid, draft.UUID)
	assert.Equal(t, text, draft.Content)
}

func TestQueryContentSearch(t *testing.T) {
	tag := rand()
	needle := "UNIQUE_SEARCH_" + rand()

	a := Create("has the needle "+needle+" inside", CreateOptions{Tags: []string{tag}})
	b := Create("no match here", CreateOptions{Tags: []string{tag}})
	defer func() {
		Trash(a)
		Trash(b)
	}()

	// Search with content filter — should only return draft "a"
	results := Query(needle, FilterInbox, QueryOptions{Tags: []string{tag}})
	uuids := getUUIDs(results)
	assert.EqualSlice(t, []string{a}, uuids)
}

func TestQueryFlagged(t *testing.T) {
	tag := rand()

	a := Create("flagged draft", CreateOptions{Tags: []string{tag}, Flagged: true})
	b := Create("unflagged draft", CreateOptions{Tags: []string{tag}})
	defer func() {
		Trash(a)
		Trash(b)
	}()

	// FilterFlagged should only return the flagged draft
	results := Query("", FilterFlagged, QueryOptions{Tags: []string{tag}})
	uuids := getUUIDs(results)
	assert.EqualSlice(t, []string{a}, uuids)
}

// ---- Workspace tests --------------------------------------------------------

func TestQueryWorkspace(t *testing.T) {
	results := QueryWorkspace("", FilterAll, QueryOptions{})
	// Empty workspace name should return empty slice, not nil
	if results == nil {
		t.Errorf("expected non-nil result from QueryWorkspace with empty name")
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results for empty workspace name, got %d", len(results))
	}
}

func TestCurrentWorkspace(t *testing.T) {
	ws := CurrentWorkspace()
	// Should return some string (Drafts always has a workspace)
	if ws == "" {
		t.Errorf("expected non-empty current workspace name")
	}
}

func TestWorkspaces(t *testing.T) {
	workspaces := Workspaces()
	if workspaces == nil {
		t.Errorf("expected non-nil workspaces list")
	}
	if len(workspaces) == 0 {
		t.Errorf("expected at least one workspace")
	}
}

// ---- Helpers ----------------------------------------------------------------

// Return a random string.
func rand() string {
	return fmt.Sprint(time.Now().UnixNano())
}

// Extract UUIDs from a slice of Drafts
func getUUIDs(ds []Draft) []string {
	uuids := make([]string, len(ds))
	for i := range ds {
		uuids[i] = ds[i].UUID
	}
	return uuids
}
