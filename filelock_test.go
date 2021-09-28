package filelock

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/charlieparkes/go-fixtures"
	"github.com/stretchr/testify/assert"
)

func TestFilelock(t *testing.T) {
	lockPath := filepath.Join(fixtures.FindPath("testdata"), fmt.Sprintf("%v.lock", fixtures.GenerateString()))

	// Does not exist
	exists, err := Exists(lockPath)
	assert.NoError(t, err)
	assert.False(t, exists)

	// Lock
	l, err := Lock(lockPath)
	assert.NoError(t, err)
	assert.Equal(t, lockPath, l.path)

	// Check lock exists
	exists, err = Exists(lockPath)
	assert.NoError(t, err)
	assert.True(t, exists)

	// Fail to lock while locked
	l2, err := Lock(lockPath)
	if l2 != nil && l2.file != nil { // Cleanup if test fails.
		defer l.Unlock()
	}
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrAlreadyLocked)

	// Check lock still exists
	exists, err = Exists(lockPath)
	assert.NoError(t, err)
	assert.True(t, exists)

	// Unlock
	assert.NoError(t, l.Unlock())

	// Check lock does not exist
	exists, err = Exists(lockPath)
	assert.NoError(t, err)
	assert.False(t, exists)
}
