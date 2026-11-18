# Slopstack Design review guide

Use the project's existing component library and tokens when supplied. Record the source and version of supplied metadata; do not claim to have inspected a library when only a screenshot was provided.

## Foundations

- Prefer the project's typography, color, spacing, and component tokens. When absent, use a restrained neutral palette, one accent, a clear type scale, and a consistent 4px spacing base.
- Prefer shadcn-style semantics: accessible controls, visible focus, clear disabled/loading states, and composable primitives.
- Keep hierarchy compact and functional. Use whitespace and alignment before adding containers.
- Keep motion purposeful and brief; support reduced-motion preferences.

## Reusable patterns

- **Review header:** project, session, status, and a clear primary review action.
- **Proposal comparison:** aligned screenshots/previews, concise direction name, rationale, and revision marker.
- **Feedback:** anchored or proposal-level comments with a severity and requested change.
- **Decision:** approve, request changes, or reject with rationale; retain prior decisions in history.

## Anti-slop checklist

- [ ] Avoid gratuitous cards and nested cards; group content with spacing and rules.
- [ ] Avoid bento layouts without a real information hierarchy need.
- [ ] Avoid huge hero headings in practical product screens.
- [ ] Avoid glow gradients and decorative gradient backgrounds.
- [ ] Avoid excessively rounded buttons; use restrained radii.
- [ ] Avoid generic three-column feature sections.
- [ ] Use icons only when they improve recognition or scanning.
- [ ] Avoid gratuitous motion; honor reduced-motion settings.
- [ ] Check contrast, keyboard focus, narrow screens, and empty/loading/error states.
- [ ] Verify the proposal fits the supplied project context and component conventions.
