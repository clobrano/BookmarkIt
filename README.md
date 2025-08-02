# BookmarkIt

**BookmarkIt** is a simple yet powerful command-line tool for managing both URLs and text snippets. Built in Go, it's designed to help you quickly save and retrieve frequently used bash commands, web links, or any other text you want to keep handy.

It integrates seamlessly with popular command-line tools like `YAD` for a simple graphical interface and `FZF` for fast, fuzzy searching. When you select an entry, `BookmarkIt` is smart enough to know what to do next: URLs are opened in your default browser, while text snippets (like bash commands) are automatically copied to your clipboard, ready to be pasted.

-----

## Features

  * **Quickly Add Bookmarks**: Use a simple graphical dialog (`YAD`) to add new entries with a key and a value. It even pre-fills the value field with content from your clipboard.
  * **Fuzzy Find and Select**: Use `FZF` to quickly search and select your saved bookmarks with fuzzy matching.
  * **Intelligent Actions**:
      * **URLs** are automatically opened in your default browser.
      * **Text snippets** (like bash commands) are copied to your system clipboard for immediate use.
  * **Simple Storage**: All bookmarks are saved in a human-readable YAML file, which you can easily edit or back up.

-----

## Prerequisites

`BookmarkIt` relies on a few external command-line tools to function. Make sure these are installed on your system:

  * **YAD** (Yet Another Dialog): Provides the graphical dialog for adding new bookmarks.
  * **FZF** (Fuzzy Finder): Powers the fuzzy search interface for finding bookmarks.
  * **Clipboard Tools**: For clipboard interaction. On Wayland, `wl-clipboard` is used, with `xclip` as a fallback.
  * **URL Opener**: `xdg-open` is used to open URLs in your default browser.

-----

## Installation

### From Source

```bash
go install github.com/clobrano/BookmarkIt@latest
```

-----

## Usage

### Add a new entry

Use the `--action add` flag to open the `YAD` dialog. If you have a URL or command in your clipboard, it will automatically populate the "link" field.

```bash
BookmarkIt --action add
```

### Find and use an entry

Use the `--action find` flag to bring up the `FZF` search interface. Start typing to filter your bookmarks, then select an entry to either open the URL or copy the text to your clipboard.

```bash
BookmarkIt --action find
```

You can also provide a starting query for `FZF`:

```bash
BookmarkIt --action find --query "git"
```

### Specify a custom bookmark file

By default, bookmarks are stored in `~/.config/BookmarkIt/bookmarks.yaml`. If you want to use a different file, use the `--file` flag.

```bash
BookmarkIt --action add --file /path/to/my/custom/bookmarks.yaml
BookmarkIt --action find --file /path/to/my/custom/bookmarks.yaml
```

-----

## Configuration

Bookmarks are stored in a YAML file. The default location is `~/.config/BookmarkIt/bookmarks.yaml`. This file is created automatically the first time you add a bookmark.

Here's an example of the file's structure:

```yaml
bookmarks:
  - key: google
    link: https://www.google.com
  - key: clone_repo
    link: git clone https://github.com/myuser/myrepo.git
```
