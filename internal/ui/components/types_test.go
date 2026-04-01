package components

import (
	"image"
	"testing"
)

func TestBaseComponent(t *testing.T) {
	comp := NewBaseComponent("test")

	if comp.Name() != "test" {
		t.Errorf("Name() = %s, want test", comp.Name())
	}

	if !comp.IsVisible() {
		t.Error("Expected component to be visible by default")
	}

	comp.SetVisible(false)
	if comp.IsVisible() {
		t.Error("Expected component to be hidden")
	}
}

func TestBaseComponentSize(t *testing.T) {
	comp := NewBaseComponent("test")

	comp.SetSize(80, 24)
	w, h := comp.Size()

	if w != 80 || h != 24 {
		t.Errorf("Size() = (%d, %d), want (80, 24)", w, h)
	}
}

func TestContainer(t *testing.T) {
	container := NewContainer("root")

	comp1 := NewBaseComponent("comp1")
	comp2 := NewBaseComponent("comp2")

	container.AddComponent(comp1)
	container.AddComponent(comp2)

	if container.GetComponent("comp1") == nil {
		t.Error("Expected to find comp1")
	}

	if container.GetComponent("comp2") == nil {
		t.Error("Expected to find comp2")
	}

	if container.GetComponent("comp3") != nil {
		t.Error("Expected comp3 not to exist")
	}
}

func TestContainerRemove(t *testing.T) {
	container := NewContainer("root")

	comp1 := NewBaseComponent("comp1")
	comp2 := NewBaseComponent("comp2")

	container.AddComponent(comp1)
	container.AddComponent(comp2)

	container.RemoveComponent("comp1")

	if container.GetComponent("comp1") != nil {
		t.Error("Expected comp1 to be removed")
	}

	if container.GetComponent("comp2") == nil {
		t.Error("Expected comp2 to still exist")
	}
}

func TestComponentSet(t *testing.T) {
	set := NewComponentSet()

	comp1 := NewBaseComponent("comp1")
	comp2 := NewBaseComponent("comp2")

	set.Add("comp1", comp1)
	set.Add("comp2", comp2)

	if set.Count() != 2 {
		t.Errorf("Count() = %d, want 2", set.Count())
	}

	if set.Get("comp1") == nil {
		t.Error("Expected to find comp1")
	}

	set.Remove("comp1")

	if set.Count() != 1 {
		t.Errorf("Count() after remove = %d, want 1", set.Count())
	}

	if set.Get("comp1") != nil {
		t.Error("Expected comp1 to be removed")
	}
}

func TestComponentSetClear(t *testing.T) {
	set := NewComponentSet()

	set.Add("comp1", NewBaseComponent("comp1"))
	set.Add("comp2", NewBaseComponent("comp2"))
	set.Add("comp3", NewBaseComponent("comp3"))

	set.Clear()

	if set.Count() != 0 {
		t.Errorf("Count() after clear = %d, want 0", set.Count())
	}
}

func TestComponentSetList(t *testing.T) {
	set := NewComponentSet()

	comp1 := NewBaseComponent("comp1")
	comp2 := NewBaseComponent("comp2")

	set.Add("comp1", comp1)
	set.Add("comp2", comp2)

	list := set.List()

	if len(list) != 2 {
		t.Errorf("List() length = %d, want 2", len(list))
	}

	found := 0
	for _, comp := range list {
		if comp.Name() == "comp1" || comp.Name() == "comp2" {
			found++
		}
	}

	if found != 2 {
		t.Error("Expected to find both components in list")
	}
}

func TestContainerUpdate(t *testing.T) {
	container := NewContainer("root")

	comp1 := NewBaseComponent("comp1")
	comp2 := NewBaseComponent("comp2")

	container.AddComponent(comp1)
	container.AddComponent(comp2)

	// Update should not error even with empty message
	updated, cmd := container.Update(nil)

	if updated == nil {
		t.Error("Expected updated container")
	}

	// cmd should be nil if no commands returned
	if cmd != nil {
		t.Error("Expected nil command")
	}
}

func TestComponentVisibility(t *testing.T) {
	comp := NewBaseComponent("test")

	tests := []struct {
		name    string
		visible bool
	}{
		{"visible", true},
		{"hidden", false},
		{"visible again", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp.SetVisible(tt.visible)
			if comp.IsVisible() != tt.visible {
				t.Errorf("SetVisible(%v), IsVisible() = %v",
					tt.visible, comp.IsVisible())
			}
		})
	}
}

func TestImagePoint(t *testing.T) {
	// Test that we can use image.Point for cursor positions
	pos := image.Pt(10, 5)

	if pos.X != 10 || pos.Y != 5 {
		t.Errorf("Point = (%d, %d), want (10, 5)", pos.X, pos.Y)
	}
}
