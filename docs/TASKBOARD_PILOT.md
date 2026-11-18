# Taskboard pilot

Taskboard is the first Slopstack Design consumer. Use its repository URL when registering the project. This starter profile records the current shadcn conventions and CSS variables read from the Taskboard repository; refresh it from source before each redesign session.

## Agent sequence

1. Sign in as a project editor and call `register_project` with the Taskboard repository URL and its current style metadata, or select the existing project with `list_projects`.
2. Call `create_design_session` with a goal and agent-observed component, token, screenshot, or preview context.
3. Call `read_design_context` before generating proposals so existing project metadata and session context are included in the design brief.
4. Submit at least two distinct `submit_proposal` calls. Include a preview URL or upload a PNG, JPEG, or WebP screenshot through the preview endpoint first.
5. Wait for the human decision with `read_reviews`, then revise the proposal in the same session with `revision_of` pointing to the prior proposal.
6. Implement the approved design in the Taskboard repository. Report the resulting commit or pull request in the proposal description; approval alone never changes repository files.

See [the Taskboard starter profile](templates/taskboard.json) and [the review guide](STYLE_GUIDE.md).
