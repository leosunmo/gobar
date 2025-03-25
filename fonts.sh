#!/bin/bash

set -ex

mkdir -p  ~/.config/gobar/fonts
mkdir -p ~/.fonts

# Material Design Symbols
wget -qO ~/.config/gobar/fonts/MaterialSymbolsOutlined.codepoints https://raw.githubusercontent.com/google/material-design-icons/refs/heads/master/variablefont/MaterialSymbolsOutlined%5BFILL%2CGRAD%2Copsz%2Cwght%5D.codepoints
wget -qO ~/.fonts/MaterialSymbolsOutlined.ttf https://raw.githubusercontent.com/google/material-design-icons/refs/heads/master/variablefont/MaterialSymbolsOutlined%5BFILL%2CGRAD%2Copsz%2Cwght%5D.ttf


# Refresh font cache
fc-cache -fv
