# go-Fluent

A **Go** port of [PyQt-Fluent-Widgets](https://github.com/zhiyiYo/PyQt-Fluent-Widgets), built on [miqt](https://github.com/mappu/miqt) (Qt5 bindings for Go). It brings the fluent, modern Windows-11-style widget library to Go so you can write desktop applications with the same look and feel as PyQt-Fluent-Widgets — without Python.

> ⚙️ **Generated with DeepSeek.** Most of the source code, documentation, and migration work in this repository was produced by the DeepSeek model.

> 📄 **中文说明见 [README.zh-CN.md](README.zh-CN.md).**

> ⚠️ This project is still in early development. Not recommended for production use.

## About

go-Fluent ports the fluent widget set of PyQt-Fluent-Widgets to Go + Qt5, including buttons, cards, inputs, lists, menus, navigation panels, setting cards, date/time pickers, dialogs, and layout helpers. Each control mirrors the original Python API as closely as the language allows, while adapting to Go idioms (embedded `*qt.QWidget` fields, callback fields, and `NewXxx` constructors).

### References

| Project | Role | License |
| --- | --- | --- |
| [zhiyiYo/PyQt-Fluent-Widgets](https://github.com/zhiyiYo/PyQt-Fluent-Widgets) | The original widget library this project ports (API & visual reference) | [GPL-3.0](https://www.gnu.org/licenses/gpl-3.0.html) |
| [mappu/miqt](https://github.com/mappu/miqt) | Go bindings for Qt that this project is built on | [MIT](https://opensource.org/licenses/MIT) |

## Features

- **Fluent design** — rounded cards, accent-color buttons, animated toggles, pill switches, and Windows-11-style navigation.
- **Broad widget coverage** — 100+ controls across `widgets`, `navigation`, `dialog_box`, `settings`, `date_time`, and `layout` packages.

## Directory structure

```
gofluent/
├── common/          # theme, style sheet, font, icon, config
├── components/      # component library
│   ├── widgets/     # base widgets
│   ├── navigation/  # navigation
│   ├── dialog_box/  # dialogs
│   ├── settings/    # setting cards
│   ├── date_time/   # date/time pickers
│   └── layout/      # layouts
├── examples/        # 80+ runnable examples
├── resources/       # QSS themes, icons, fonts
└── docs/            # usage docs
```

## Build

This project builds on top of [miqt](https://github.com/mappu/miqt), so the Qt toolchain and cross-compilation environment follow miqt's official guide. See **[miqt › Building](https://github.com/mappu/miqt#building)** for how to install Qt5 and set up Windows cross-compilation (MXE / Docker / MSYS2).

Once the toolchain is ready, use the scripts in this repository:

```bash
cd gofluent

# Build the gallery demo
bash build_exes.sh

# Build EVERY example, each .exe placed next to its own main.go
bash build_all_examples.sh
```

## Documentation

- [控件使用文档（中文）](docs/控件使用文档.md) — 每个控件的功能、参数、调用方法与效果。
- [Controls Usage Guide (English)](docs/controls_usage.md) — function, parameters, usage and effect for every control.

## License

**GPL-3.0** — see [LICENSE](LICENSE).

This project is a port of PyQt-Fluent-Widgets, which is licensed under **GPL-3.0**. GPL-3.0 is the license that is compatible with *both* upstream projects: the GPL requires derivative works to stay under the same license, while the permissive MIT license of miqt allows its code to be incorporated into a GPL-3.0 project.
