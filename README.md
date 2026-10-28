# django-launch

A terminal wizard that scaffolds a Django project. Answer seven questions and it creates a folder with a virtual environment, Django 6.1 installed, the settings and URLs patched for the packages you picked, starter apps and a Dockerfile if you want one.

## what it creates

```
<project name>/
├── .venv/                  uv venv, or python -m venv
├── requirements.txt        Django>=6.1 plus your selections
├── manage.py
├── .env                    SECRET_KEY and DB variables (django-environ only)
├── .gitignore              keeps .env and media out of version control
├── <project name>/
│   ├── settings.py         patched for your selections
│   ├── urls.py             patched for your selections and app routes
│   ├── wsgi.py
│   ├── asgi.py
│   └── __init__.py
├── <app>/                  one folder per selected starter app (api, base)
│   ├── urls.py
│   └── templates/          only with per-app template folders
├── templates/              only with a root templates folder
├── Dockerfile              only if you answered yes
└── .dockerignore
```

The project is created in place. The tool does not run `migrate` or `runserver`; the final screen prints those commands.

## requirements

- Go 1.26.5 or newer, to build it.
- `uv` or `python` on your PATH. The app prefers uv when it finds it and falls back to `python -m venv` plus pip. If neither is installed, the generation step fails on the error screen with the reason.
- Network access, because the generator installs packages with `uv pip install` or `pip install`.

## build and run

```
go build -o django-launch
./django-launch
```

On Windows the binary is `django-launch.exe`. Run it from the directory where you want the project folder to appear, since it creates `./<project name>`.

The wizard runs in the alternate screen. The ASCII banner renders in full when it fits the terminal and shrinks to a one-line title on short terminals so nothing gets clipped.

## the seven steps

**1. Project name.** Anything that is a valid Python identifier. Empty names, leading digits, hyphens, spaces, and reserved words such as `django` or `admin` are rejected on the spot.

**2. Docker support.** Yes writes a `Dockerfile` on `python:3.12-slim` and a `.dockerignore`. No skips both.

**3. Dependencies.** Nine optional packages, listed below.

**4. Databases.** PostgreSQL, Sqlite, both, or neither. A selection rewrites the `DATABASES` block in `settings.py` to read connection values from environment variables, and PostgreSQL adds `psycopg[binary]` to `requirements.txt`.

**5. Starter apps.** Ask whether to create apps, then pick from `api` and `base` checkboxes. Each app is created with `manage.py startapp`, gets its own `urls.py`, and a route in the project's `urls.py`.

**6. Templates folder.** No templates folder, a single root `templates/` folder, or a `templates/` folder inside each created app (only offered when apps were created). Folders are created empty, with a `.gitkeep` so they survive in git.

**7. Confirmation.** Shows the project name, target path, Docker choice, dependency list, database, apps, templates, and whether the install will run through uv or pip. Enter starts the generation.

### keys

| step | keys |
| --- | --- |
| project name | type or paste, `enter` to continue, `ctrl+c` to quit |
| docker | `up`/`down`/`left`/`right` or `h`/`j`/`k`/`l` to choose, `y` selects Yes and continues, `n` selects No and continues, `enter` continues, `esc` or `backspace` goes back |
| dependencies | `up`/`down` or `k`/`j` moves with wraparound, `space` or `x` toggles the row, `a` toggles every row, `enter` continues, `esc` or `backspace` goes back |
| databases | `up`/`down` or `k`/`j` moves with wraparound, `space` or `x` toggles the row, `a` toggles every row, `enter` continues, `esc` or `backspace` goes back |
| starter apps | ask phase: `up`/`down` or `h`/`j`/`k`/`l` to choose, `y` answers Yes and opens the app list, `n` answers No and skips ahead, `enter` continues, `esc` or `backspace` goes back. select phase: `up`/`down` or `k`/`j` moves, `space` or `x` toggles the row, `a` toggles every row, `enter` continues, `esc` or `backspace` returns to the question |
| templates | `up`/`down` or `h`/`j`/`k`/`l` to choose, `enter` continues, `esc` or `backspace` goes back |
| confirm | `enter` or `y` starts, `esc`/`backspace`/`b` goes back, `q` or `n` cancels |
| generating | `q`, `n`, or `esc` cancels, `ctrl+c` quits |
| done or error | `enter`, `q`, or `esc` exits |

Each step prints a hint line with its keys, and that hint matches this table. `ctrl+c` quits from any step.

## optional packages

| package | what it does |
| --- | --- |
| djangorestframework | toolkit for building Web APIs |
| pillow | image handling, needed for `ImageField` and media |
| django-silk | profiling and SQL query inspection |
| djangorestframework-simplejwt | token authentication for the REST framework |
| django-filter | filtering for API views |
| django-cors-headers | CORS response headers |
| django-environ | `.env` support: creates `.env` (with the project's `SECRET_KEY`), `.gitignore`, and wires settings |
| django-debug-toolbar | debug panels during development |
| django-modeltranslation | translated model fields |

Seven of these get `INSTALLED_APPS`, middleware, or URL changes written into `settings.py` and `urls.py`. Pillow also adds `MEDIA_URL`, `MEDIA_ROOT`, and the static media route. `django-environ` additionally generates `.env` and `.gitignore` and moves `SECRET_KEY` and database values into environment lookups.

## development

```
go test ./...          # full suite, about 10 seconds
go test -short ./...   # skips the end-to-end test, about 2 seconds
gofmt -l .
go vet ./...
```

The end-to-end test builds a real Django project in a temp directory, so it needs Python or uv and network access. Everything else runs offline.

Continuous integration lives in `.github/workflows/ci.yml`. Its automatic triggers are commented out, so nothing runs until you uncomment them.

## license

MIT. See [LICENSE](LICENSE).
