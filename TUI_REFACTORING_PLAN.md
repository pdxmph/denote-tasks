# TUI Refactoring Implementation Plan

## Overview

This plan outlines the step-by-step implementation of the TUI refactoring improvements created in the previous commit. The goal is to gradually integrate the new helpers while maintaining functionality and testing at each step.

## Implementation Phases

### Phase 1: Constants Integration (Low Risk, High Impact)
**Timeline**: 1-2 hours
**Files affected**: All TUI files

1. **Replace magic strings with constants**
   - [ ] Replace status symbols ("○", "✓", etc.) with StatusSymbol* constants
   - [ ] Replace priority strings ("p1", "p2", "p3") with PriorityLevels
   - [ ] Replace error messages with constants from constants.go
   - [ ] Replace magic numbers (6, 10, etc.) with named constants

2. **Update field references**
   - [ ] Convert string-based editingField to EditableField type
   - [ ] Convert sortBy strings to SortField type

3. **Testing checkpoint**
   - [ ] Build and run basic TUI operations
   - [ ] Verify sorting, filtering, and field editing still work

### Phase 2: Navigation Handler Integration (Medium Risk, High Impact)
**Timeline**: 2-3 hours
**Files affected**: keys.go, project_view_keys.go, task_view_keys.go

1. **Start with simple cases**
   - [ ] Replace handlePreviewKeys navigation logic
   - [ ] Test preview scrolling functionality
   - [ ] Replace handleProjectSelectKeys navigation logic
   - [ ] Test project selection navigation

2. **Move to complex cases**
   - [ ] Replace handleNormalKeys navigation (main list)
   - [ ] Replace handleProjectViewKeys navigation
   - [ ] Replace handleTaskViewKeys navigation

3. **Testing checkpoint**
   - [ ] Test all navigation keys (j/k, g/G, ctrl+d/ctrl+u)
   - [ ] Verify cursor bounds checking
   - [ ] Test with empty lists and single items

### Phase 3: Field Renderer Integration (Medium Risk, Medium Impact)
**Timeline**: 3-4 hours
**Files affected**: task_view.go, project_view.go, views.go

1. **Task view rendering**
   - [ ] Replace renderField implementations in task_view.go
   - [ ] Use FieldRenderer for all task metadata display
   - [ ] Test task view with various field states

2. **Project view rendering**
   - [ ] Replace renderField implementations in project_view.go
   - [ ] Ensure consistent styling with task view
   - [ ] Test project view display

3. **List view field rendering**
   - [ ] Update renderTaskLine to use consistent field rendering
   - [ ] Update renderProjectLine similarly
   - [ ] Test list displays

### Phase 4: Mode Handler Refactoring (High Risk, Medium Impact)
**Timeline**: 4-5 hours
**Files affected**: model.go, keys.go

1. **Create handler map structure**
   - [ ] Define modeHandler interface
   - [ ] Create handler map with existing functions
   - [ ] Replace main switch statement in handleKeyPress

2. **Refactor View() method**
   - [ ] Create render map similar to handler map
   - [ ] Replace switch statement in View()
   - [ ] Test all modes still render correctly

3. **Testing checkpoint**
   - [ ] Test mode transitions
   - [ ] Verify all modes accessible and functional
   - [ ] Check for any missing key handlers

### Phase 5: State Management Refactoring (High Risk, Low Priority)
**Timeline**: 6-8 hours
**Files affected**: model.go

1. **Split Model struct**
   - [ ] Create FileManager for file operations
   - [ ] Create UIState for display state
   - [ ] Create FilterState for filter management
   - [ ] Create EditorState for editing operations

2. **Update method receivers**
   - [ ] Update methods to work with new sub-structs
   - [ ] Maintain backward compatibility during transition
   - [ ] Gradually migrate functionality

3. **Full regression testing**
   - [ ] Test all functionality thoroughly
   - [ ] Performance testing
   - [ ] Memory usage verification

## Testing Strategy

### Unit Tests (Create if not existing)
```go
// navigation_test.go
func TestNavigationHandler(t *testing.T) {
    nav := NewNavigationHandler(10, false)
    // Test boundary conditions
    // Test wrap behavior
    // Test page navigation
}

// field_renderer_test.go
func TestFieldRenderer(t *testing.T) {
    renderer := NewFieldRenderer()
    // Test each render method
    // Test edit mode rendering
    // Test empty value handling
}
```

### Integration Tests
1. Create test task files in a temp directory
2. Run TUI with test config
3. Verify:
   - Navigation works in all views
   - Field editing preserves data
   - Sorting/filtering unchanged
   - File operations work correctly

### Manual Testing Checklist
- [ ] Launch TUI and navigate through all modes
- [ ] Create, edit, and delete tasks
- [ ] Test all keyboard shortcuts
- [ ] Test with various terminal sizes
- [ ] Test with empty task directory
- [ ] Test with 100+ tasks (performance)

## Rollback Plan

Each phase should be implemented in a separate commit. If issues arise:

1. **Immediate rollback**: `git revert HEAD`
2. **Partial rollback**: Revert specific phase commits
3. **Feature flag**: Add config option to use old/new behavior

## Success Metrics

1. **Code Quality**
   - Reduced line count by 20-30%
   - Eliminated code duplication
   - Improved test coverage

2. **Performance**
   - No degradation in response time
   - Memory usage stable or improved
   - Smooth scrolling with 1000+ items

3. **Maintainability**
   - New features easier to add
   - Bugs easier to locate and fix
   - Code more self-documenting

## Next Steps After Completion

1. **Documentation**
   - Update code comments
   - Create developer guide for TUI
   - Document new patterns for contributors

2. **Additional Improvements**
   - Add keyboard macro support
   - Implement undo/redo
   - Add customizable key bindings
   - Create theme system using constants

3. **Performance Optimizations**
   - Lazy loading of task metadata
   - Virtual scrolling for large lists
   - Background file watching

## Risk Mitigation

1. **Always test after each sub-phase**
2. **Keep refactoring commits atomic**
3. **Don't refactor during feature development**
4. **Have another developer review changes**
5. **Test on multiple platforms (Linux, macOS, Windows)**

## Estimated Total Time

- Phase 1: 1-2 hours ✓ Low risk, start here
- Phase 2: 2-3 hours ✓ High value
- Phase 3: 3-4 hours ✓ User-visible improvements
- Phase 4: 4-5 hours ⚠️  Careful testing needed
- Phase 5: 6-8 hours ⚠️  Consider deferring

**Total: 16-22 hours of focused work**

## Implementation Order Recommendation

1. Start with Phase 1 (constants) - immediate value, low risk
2. Then Phase 2 (navigation) - eliminates most duplication
3. Then Phase 3 (field rendering) - visible improvement
4. Evaluate before proceeding to Phase 4 & 5 - higher risk

This incremental approach allows for value delivery at each phase while maintaining a working application throughout the refactoring process.