#!/bin/bash
cd /Users/yoeven/Documents/GitHub/weaviate
go build -v ./modules/multi2vec-jigsawstack/... > build_log.txt 2>&1
echo "Build completed. Check build_log.txt for details."
if [ $? -eq 0 ]; then
  echo "Build successful!"
else
  echo "Build failed. Check build_log.txt for errors."
  cat build_log.txt
  echo ""
  echo "Trying to fix by running go mod tidy..."
  go mod tidy
  go build -v ./modules/multi2vec-jigsawstack/... > build_log.txt 2>&1
  if [ $? -eq 0 ]; then
    echo "Build successful after go mod tidy!"
  else
    echo "Build still failed after go mod tidy. Check build_log.txt for errors."
    cat build_log.txt
  fi
fi 