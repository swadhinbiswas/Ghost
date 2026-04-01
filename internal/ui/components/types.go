package components

import (
	"image"

	tea "charm.land/bubbletea/v2"
	uv "github.com/charmbracelet/ultraviolet"
)

// Component is the interface all UI components must implement.
// Components are responsible for rendering a specific part of the UI
// within their assigned rectangle.
type Component interface {
	// Draw renders the component within the provided rectangle on the screen.
	// The component should only render within the bounds of the rectangle.
	Draw(scr uv.Screen, area uv.Rectangle)

	// Update processes messages for the component.
	// Returns the updated component and any commands to execute.
	Update(msg tea.Msg) (Component, tea.Cmd)

	// Name returns the component's unique identifier.
	Name() string
}

// DrawableComponent is a component that can also report its state for debugging.
type DrawableComponent interface {
	Component

	// Info returns information about the component's current state.
	Info() ComponentInfo
}

// ComponentInfo contains debugging/info about a component.
type ComponentInfo struct {
	Name      string
	Height    int
	Width     int
	LineCount int
	IsVisible bool
}

// ScrollableComponent is a component that supports scrolling.
type ScrollableComponent interface {
	Component

	// ScrollUp scrolls up by n lines.
	ScrollUp(n int)

	// ScrollDown scrolls down by n lines.
	ScrollDown(n int)

	// ScrollToTop scrolls to the very top.
	ScrollToTop()

	// ScrollToBottom scrolls to the very bottom.
	ScrollToBottom()

	// IsAtTop returns whether the component is scrolled to the top.
	IsAtTop() bool

	// IsAtBottom returns whether the component is scrolled to the bottom.
	IsAtBottom() bool
}

// FocusableComponent is a component that can receive focus.
type FocusableComponent interface {
	Component

	// Focus sets focus to this component.
	Focus()

	// Blur removes focus from this component.
	Blur()

	// IsFocused returns whether the component has focus.
	IsFocused() bool
}

// EditableComponent is a component that allows text editing.
type EditableComponent interface {
	Component

	// SetText sets the component's text content.
	SetText(text string)

	// GetText returns the component's text content.
	GetText() string

	// Clear clears the component's content.
	Clear()

	// CursorPosition returns the current cursor position (row, col).
	CursorPosition() image.Point
}

// Component Event Messages

// ComponentFocusedMsg indicates a component gained focus.
type ComponentFocusedMsg struct {
	ComponentName string
}

// ComponentBlurredMsg indicates a component lost focus.
type ComponentBlurredMsg struct {
	ComponentName string
}

// ComponentScrolledMsg indicates a component was scrolled.
type ComponentScrolledMsg struct {
	ComponentName string
	Direction     string // "up" or "down"
	Lines         int
}

// ComponentUpdatedMsg indicates a component's content changed.
type ComponentUpdatedMsg struct {
	ComponentName string
	Data          interface{}
}

// ComponentErrorMsg indicates an error in a component.
type ComponentErrorMsg struct {
	ComponentName string
	Error         error
}

// BaseComponent provides common functionality for components.
type BaseComponent struct {
	name    string
	visible bool
	height  int
	width   int
}

// NewBaseComponent creates a new base component.
func NewBaseComponent(name string) *BaseComponent {
	return &BaseComponent{
		name:    name,
		visible: true,
		height:  0,
		width:   0,
	}
}

// Name implements Component interface.
func (c *BaseComponent) Name() string {
	return c.name
}

// SetVisible sets whether the component is visible.
func (c *BaseComponent) SetVisible(visible bool) {
	c.visible = visible
}

// IsVisible returns whether the component is visible.
func (c *BaseComponent) IsVisible() bool {
	return c.visible
}

// SetSize sets the component's size (for layout calculations).
func (c *BaseComponent) SetSize(width, height int) {
	c.width = width
	c.height = height
}

// Size returns the component's size.
func (c *BaseComponent) Size() (width, height int) {
	return c.width, c.height
}

// Draw is a no-op in base component (override in subclasses).
func (c *BaseComponent) Draw(scr uv.Screen, area uv.Rectangle) {
	// Override in subclass
}

// Update is a no-op in base component (override in subclasses).
func (c *BaseComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	return c, nil
}

// Container is a component that holds other components.
type Container struct {
	*BaseComponent
	components []Component
	layout     image.Rectangle
}

// NewContainer creates a new container.
func NewContainer(name string) *Container {
	return &Container{
		BaseComponent: NewBaseComponent(name),
		components:    []Component{},
	}
}

// AddComponent adds a component to the container.
func (c *Container) AddComponent(comp Component) {
	c.components = append(c.components, comp)
}

// RemoveComponent removes a component by name.
func (c *Container) RemoveComponent(name string) {
	for i, comp := range c.components {
		if comp.Name() == name {
			c.components = append(c.components[:i], c.components[i+1:]...)
			return
		}
	}
}

// GetComponent returns a component by name.
func (c *Container) GetComponent(name string) Component {
	for _, comp := range c.components {
		if comp.Name() == name {
			return comp
		}
	}
	return nil
}

// Draw draws all child components.
func (c *Container) Draw(scr uv.Screen, area uv.Rectangle) {
	if !c.visible {
		return
	}

	for _, comp := range c.components {
		comp.Draw(scr, area)
	}
}

// Update updates all child components.
func (c *Container) Update(msg tea.Msg) (Component, tea.Cmd) {
	var cmds []tea.Cmd

	for i, comp := range c.components {
		updated, cmd := comp.Update(msg)
		c.components[i] = updated
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	if len(cmds) == 0 {
		return c, nil
	}

	return c, tea.Batch(cmds...)
}

// ComponentSet is a set of named components for easy lookup.
type ComponentSet struct {
	components map[string]Component
}

// NewComponentSet creates a new component set.
func NewComponentSet() *ComponentSet {
	return &ComponentSet{
		components: make(map[string]Component),
	}
}

// Add adds a component to the set.
func (cs *ComponentSet) Add(name string, comp Component) {
	cs.components[name] = comp
}

// Get retrieves a component by name.
func (cs *ComponentSet) Get(name string) Component {
	return cs.components[name]
}

// Remove removes a component by name.
func (cs *ComponentSet) Remove(name string) {
	delete(cs.components, name)
}

// List returns all components.
func (cs *ComponentSet) List() []Component {
	components := make([]Component, 0, len(cs.components))
	for _, comp := range cs.components {
		components = append(components, comp)
	}
	return components
}

// Count returns the number of components in the set.
func (cs *ComponentSet) Count() int {
	return len(cs.components)
}

// Clear removes all components.
func (cs *ComponentSet) Clear() {
	cs.components = make(map[string]Component)
}

// DrawAll draws all components with their assigned rectangles.
// The caller is responsible for providing the correct rectangles for each component.
type DrawLayout struct {
	Component Component
	Area      uv.Rectangle
}

// DrawAllComponents draws all layouts in sequence.
func DrawAllComponents(scr uv.Screen, layouts []DrawLayout) {
	for _, layout := range layouts {
		if layout.Component != nil {
			layout.Component.Draw(scr, layout.Area)
		}
	}
}
