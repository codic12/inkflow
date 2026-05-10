# Inkflow (Local-First)

Inkflow has been migrated to a serverless, local-first architecture. All your data is stored securely in your browser's **IndexedDB**.

## Features
- **No Server Required:** Simply open `index.html` in any modern web browser.
- **Local Storage:** Notebooks, notes, and images are saved locally to your browser's persistent storage.
- **Export/Import:** Use the Settings icon (cog) in the sidebar footer to export your entire library as a JSON file or import a backup.
- **Offline Capable:** Once loaded, the app works entirely offline.

## Running the App
1. Open `index.html` in Chrome, Firefox, Safari, or Edge.
2. Start drawing!

## Migration from Server Version
If you have data from the old server-side version, you can import your `meta.json` and note contents using the "Import Library" feature in the Settings menu (requires the data to be in the expected JSON format).
