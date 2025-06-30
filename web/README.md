<h1 align="center">Crack Hash - UI</h1>

Web interface for Crack-HASH

## Configuration

Configuration is done via JSON files [`public/assets/config/appConfig.json`](public/assets/config/appConfig.json).

```json
{
  "managerUrl": "http://localhost:8080/api/manager"
}
```

## Usage

- `npm run dev`          - Starts the development server using [Vite](https://vitejs.dev/). Ideal for local development
with hot module replacement (HMR).
- `npm run dev:debug`    - Starts the Vite development server in debug mode, useful for inspecting or debugging the
development environment.
- `npm run build`        - Builds the application for production. This also runs type checking before building.
- `npm run preview`      - Locally previews the production build using Vite’s built-in preview server. Make sure to run
`build` first.
- `npm run serve`        - Serves the already built dist folder using Vite’s static server. Useful for testing the
production-ready assets.
- `npm run test:unit`    - Runs unit tests using [Vitest](https://vitest.dev/), a fast and modern test framework designed
for Vite projects.
- `npm run build-only`   - Builds the application without running type checking. Use this if you've already confirmed
types are correct.
- `npm run type-check`   - Performs TypeScript type checking using [vue-tsc](https://www.npmjs.com/package/vue-tsc),
ensuring your Vue code has correct types.
- `npm run lint:oxlint`  - Lints the project using [Oxlint](https://oxc.rs/guide/linter.html), an alternative Rust-based
linter for JavaScript/TypeScript.
- `npm run lint:oxlint:fix` - Same as above but automatically fixes fixable issues detected by Oxlint.
- `npm run lint:eslint`  - Lints the project using [ESLint](https://eslint.org/) for code quality and style issues.
- `npm run lint:eslint:fix` - Same as above but automatically fixes fixable issues detected by ESLint.
- `npm run lint`         - Runs all linters (`lint:oxlint`, `lint:eslint`, etc.) in sequence
using [npm-run-all](https://github.com/mysticatea/npm-run-all).
- `npm run format`       - Formats the source files in the `src/` directory using [Prettier](https://prettier.io/),
ensuring consistent code style.
