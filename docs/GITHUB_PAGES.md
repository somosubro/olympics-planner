# GitHub Pages for the privacy policy

The privacy policy for the Custom GPT lives at [`gpt/privacy-policy.html`](gpt/privacy-policy.html). Hosting the **`docs/`** folder on **GitHub Pages** gives you a stable **HTTPS** URL for ChatGPT’s **Privacy policy** field.

## One-time setup

1. Push this repository to GitHub (if it is not already there). Your repo name might be `olympics-planner` or something else — use **your** repo name in URLs below.

2. On GitHub: open the repo → **Settings** → **Pages** (under “Code and automation”).

3. Under **Build and deployment** → **Source**, choose **Deploy from a branch**.

4. **Branch:** `main` (or your default branch), **Folder:** **`/docs`**, then **Save**.

5. Wait until the banner shows “Your site is live at …” (often one or two minutes).

## Your privacy policy URL

For a **project site**, GitHub serves the contents of `docs/` at:

```text
https://<YOUR_GITHUB_USERNAME>.github.io/<REPOSITORY_NAME>/
```

The privacy policy HTML file is at **`gpt/privacy-policy.html`** inside `docs/`, so the URL to paste into ChatGPT is:

```text
https://<YOUR_GITHUB_USERNAME>.github.io/<REPOSITORY_NAME>/gpt/privacy-policy.html
```

**Example** (replace with your username and repo):

`https://octocat.github.io/olympics-planner/gpt/privacy-policy.html`

Open that link in a browser to confirm it loads (**200 OK**, green lock).

## Before you share

1. Edit **`docs/gpt/privacy-policy.html`** and replace `replace-with-your-email@example.com` with a real contact email or support link.
2. Optionally edit **`docs/gpt/privacy-policy.md`** to match.
3. Commit and push; Pages will redeploy in a minute or two.

## Troubleshooting

| Issue | What to try |
|--------|----------------|
| 404 | Confirm **Pages** source is **main** + **/docs** and the file path includes **`gpt/privacy-policy.html`**. |
| Wrong site | Project Pages URL always includes **`<repo>`** in the path: `…github.io/<repo>/…`. |
| Old content | Hard-refresh or wait for the Pages build to finish. |

No Jekyll theme is required: an empty **`docs/.nojekyll`** file disables Jekyll so files are served **as static HTML**.
