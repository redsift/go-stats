package rules_test

import (
	"testing"

	"github.com/redsift/go-stats/router/rules"
	"github.com/stretchr/testify/assert"
)

func TestByName(t *testing.T) {
	byName := rules.ByName("test")
	assert.True(t, byName.Match("test", nil), "did not match")
	assert.False(t, byName.Match("Test", nil), "did match when not expected")
}

func TestByNameFold(t *testing.T) {
	byName := rules.ByNameFold("test")
	assert.True(t, byName.Match("test", nil), "did not match")
	assert.True(t, byName.Match("Test", nil), "did not match")
	assert.False(t, byName.Match("notTest", nil), "did match when not expected")
}

func TestNot(t *testing.T) {
	byName := rules.ByName("test")
	notByName := byName.Not()
	assert.Equal(t, byName, notByName.Not())
	assert.False(t, notByName.Match("test", nil), "did not match")
	assert.True(t, notByName.Match("Test", nil), "did match when not expected")
}

func TestByTag(t *testing.T) {
	byTag := rules.ByTag("test:value")
	assert.True(t, byTag.Match("test", []string{"test:value"}), "did not match")
	assert.False(t, byTag.Match("Test", []string{"test:other"}), "did match when not expected")
	assert.False(t, byTag.Match("Test", []string{"other:value"}), "did match when not expected")
}

func TestByTagName(t *testing.T) {
	byTagName := rules.ByTagName("test")
	assert.True(t, byTagName.Match("test", []string{"test:value"}), "did not match")
	assert.True(t, byTagName.Match("Test", []string{"test:other"}), "did not match")
	assert.False(t, byTagName.Match("Test", []string{"other:value"}), "did match when not expected")
}
