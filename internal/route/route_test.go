package route

import (
	"testing"

	"hookrelay/internal/model"
)

func TestRegistryRegisterGetList(t *testing.T) {
	registry := NewRegistry()
	cb := model.NewCallback("cb-1", "order.created", "http://host/hook", "s")
	if err := registry.Register(cb); err != nil {
		t.Fatalf("register failed: %v", err)
	}
	got, ok := registry.Get("cb-1")
	if !ok || got.ID != "cb-1" {
		t.Fatal("registered callback should be retrievable")
	}
	if registry.Len() != 1 {
		t.Fatalf("expected one callback, got %d", registry.Len())
	}
	if len(registry.List()) != 1 {
		t.Fatal("List should contain the registered callback")
	}
}

func TestRegistryRejectsInvalidCallback(t *testing.T) {
	registry := NewRegistry()
	cb := model.NewCallback("cb-1", "order.created", "not-a-url", "s")
	if err := registry.Register(cb); err == nil {
		t.Fatal("invalid callback url should be rejected")
	}
}

func TestRegistryUnregister(t *testing.T) {
	registry := NewRegistry()
	cb := model.NewCallback("cb-1", "order.created", "http://host/hook", "s")
	_ = registry.Register(cb)
	if err := registry.Unregister("cb-1"); err != nil {
		t.Fatalf("unregister failed: %v", err)
	}
	if _, ok := registry.Get("cb-1"); ok {
		t.Fatal("unregistered callback should not be found")
	}
	if err := registry.Unregister("cb-1"); err == nil {
		t.Fatal("second unregister should fail")
	}
}

func TestMatchReturnsEnabledCallbacksSorted(t *testing.T) {
	registry := NewRegistry()
	second := model.NewCallback("cb-2", "order.created", "http://host/b", "s")
	first := model.NewCallback("cb-1", "order.created", "http://host/a", "s")
	disabled := model.NewCallback("cb-3", "order.created", "http://host/c", "s")
	disabled.Enabled = false
	_ = registry.Register(second)
	_ = registry.Register(first)
	_ = registry.Register(disabled)
	matched := registry.Match("order.created")
	if len(matched) != 2 {
		t.Fatalf("expected two matches, got %d", len(matched))
	}
	if matched[0].ID != "cb-1" || matched[1].ID != "cb-2" {
		t.Fatalf("matches should be sorted by id: %v, %v", matched[0].ID, matched[1].ID)
	}
}

func TestLoaderLoadsDefaults(t *testing.T) {
	registry := NewRegistry()
	loader := NewLoader(registry)
	loaded, err := loader.LoadDefaults()
	if err != nil {
		t.Fatalf("load defaults failed: %v", err)
	}
	if loaded != 3 || registry.Len() != 3 {
		t.Fatalf("expected three default callbacks, got %d", loaded)
	}
}
