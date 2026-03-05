# App Chat

Bubble Tea orchestration for chat rounds, turns, and message list behavior.

## Rules

1. Scope stream/tool events by turn ID before applying updates.
2. Keep reducers pure and handlers side-effectful.
3. Message list projection math must be shared between viewport and rendering.
4. User-cancel semantics must stay non-error and non-persisted for committed assistant output.

## Structure Expectations

1. `messagelist/update.go` should remain a router.
2. Input-family handlers live in dedicated files (`update_key`, `update_mouse`, `update_lifecycle`).
3. Transition logic belongs in reducer files.

## Testing Requirements

For behavior changes in this tree:

1. Add/adjust reducer unit tests.
2. Add/adjust behavior tests in `messagelist/behavior_test.go` for user-visible outcomes.
3. Run `go test ./internal/chat ./internal/app/chat/... -count=1`.

