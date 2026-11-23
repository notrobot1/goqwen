#!/usr/bin/env bash
# Собирает llama.cpp как статические библиотеки для текущей платформы.
# Результат кладёт в internal/llamacpp/libs/<GOOS>_<GOARCH>/
#
# Использование:
#   ./scripts/build_llama_libs.sh           # текущая платформа
#   LLAMA_CUDA=ON ./scripts/build_llama_libs.sh  # с CUDA (Linux)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
LLAMA_SRC="${LLAMA_SRC:-/tmp/llama.cpp}"
LLAMA_VERSION="${LLAMA_VERSION:-HEAD}"

GOOS="$(go env GOOS)"
GOARCH="$(go env GOARCH)"
DEST="$REPO_ROOT/internal/llamacpp/libs/${GOOS}_${GOARCH}"
mkdir -p "$DEST"

echo "==> Сборка llama.cpp для ${GOOS}/${GOARCH}"
echo "    Источник: $LLAMA_SRC"
echo "    Назначение: $DEST"

# Клонируем если нет
if [ ! -d "$LLAMA_SRC" ]; then
  echo "==> Клонируем llama.cpp..."
  git clone --depth 1 https://github.com/ggerganov/llama.cpp "$LLAMA_SRC"
fi

# Копируем заголовки
echo "==> Копируем заголовки..."
cp "$LLAMA_SRC/include/llama.h" "$REPO_ROOT/internal/llamacpp/include/"
cp "$LLAMA_SRC/ggml/include/"*.h "$REPO_ROOT/internal/llamacpp/include/"

# Настройки cmake по платформе
CMAKE_ARGS=(
  -DCMAKE_BUILD_TYPE=Release
  -DBUILD_SHARED_LIBS=OFF
  -DLLAMA_BUILD_SERVER=OFF
  -DLLAMA_BUILD_TESTS=OFF
  -DLLAMA_BUILD_EXAMPLES=OFF
  -DLLAMA_BUILD_TOOLS=OFF
  -DCMAKE_CXX_FLAGS="-w"
  -DCMAKE_C_FLAGS="-w"
)

case "${GOOS}_${GOARCH}" in
  darwin_arm64)
    CMAKE_ARGS+=(-DGGML_METAL=ON -DCMAKE_OSX_ARCHITECTURES=arm64)
    ;;
  darwin_amd64)
    CMAKE_ARGS+=(-DGGML_METAL=ON -DCMAKE_OSX_ARCHITECTURES=x86_64)
    ;;
  linux_*)
    CMAKE_ARGS+=(-DGGML_METAL=OFF)
    if [ "${LLAMA_CUDA:-OFF}" = "ON" ]; then
      CMAKE_ARGS+=(-DGGML_CUDA=ON)
    fi
    if [ "${LLAMA_VULKAN:-OFF}" = "ON" ]; then
      CMAKE_ARGS+=(-DGGML_VULKAN=ON)
    fi
    ;;
esac

BUILD_DIR="/tmp/llama_build_${GOOS}_${GOARCH}"
echo "==> cmake configure..."
cmake -S "$LLAMA_SRC" -B "$BUILD_DIR" "${CMAKE_ARGS[@]}"

JOBS=$(sysctl -n hw.logicalcpu 2>/dev/null || nproc 2>/dev/null || echo 4)
echo "==> cmake build (${JOBS} потоков)..."
cmake --build "$BUILD_DIR" --target llama ggml -j"$JOBS"

echo "==> Копируем .a файлы в $DEST..."
find "$BUILD_DIR" -name "*.a" | while read -r f; do
  base=$(basename "$f")
  cp "$f" "$DEST/$base"
  echo "    $base ($(du -sh "$f" | cut -f1))"
done

echo ""
echo "✓ Готово! Теперь можно собирать с GPU:"
echo "  CGO_ENABLED=1 go build -tags llamacpp ."
