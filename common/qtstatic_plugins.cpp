#ifdef MIQT_WINDOWSQTSTATIC

#include <QtPlugin>

// Static Qt5 image-format plugins (JPEG/ICO/GIF) are linked via CGO_LDFLAGS
// (-lqjpeg -lqico -lqgif) but miqt's cflags_windowsqtstatic.cpp only imports
// the platform/style plugins. Register these so the static runtime can decode
// JPEG/ICO/GIF images (e.g. the gallery demo's .jpg assets and the image_label
// demo's .gif asset).
Q_IMPORT_PLUGIN(QJpegPlugin)
Q_IMPORT_PLUGIN(QGifPlugin)

#endif
