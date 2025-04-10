name: Check Homebrew Formula Syntax

on:
  push:
    branches: [main, dev]
  pull_request:
    branches: [main, dev]

jobs:
  check-formulas:
    runs-on: ubuntu-latest
    name: Validate Formula Syntax

    steps:
      - name: Checkout repository
        uses: actions/checkout@v4

      - name: Install Ruby
        uses: ruby/setup-ruby@v1
        with:
          ruby-version: '3.1' # or use system ruby

      - name: Check Ruby syntax in Formula files
        run: |
          for file in Formula/*.rb; do
            echo "Checking $file..."
            ruby -c "$file"
          done