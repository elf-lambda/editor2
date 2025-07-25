# Editor

A custom graphical text editor built from the ground up, implementing font rendering and UI components from scratch using Raylib.

## Demo

![Demo Video 1](screenshots/forth.gif)


## Features

- **Custom Font Rendering**: Font bitmap generation and text rendering implemented from scratch
- **Native UI Components**: All UI elements built using raw Raylib primitives
- **Text Editing**: Full cursor navigation and text manipulation
- **Opening/Saving Files**
- **File/Directory Picker**: GUI File picker / directory navigator
- **Text selection/deletion**
- **Text copy/paste, Selection copy/paste**
- **Daily Note Taking ui options**: Allows to automatically create dd-mm-yyyy files to take notes
- **Undo/Redo Snapshots**: Ctrl+Z, Ctrl+Shift+Z to undo/redo changes made in the text

## Building

```bash
chmod +x build.sh
./build.sh
```

## Font Extraction

Inside utils folder there are 2 scripts that were used to extract individual png images from a font and then generating the bitmap array that is used in the editor.


## License

See [LICENSE](LICENSE) for details.