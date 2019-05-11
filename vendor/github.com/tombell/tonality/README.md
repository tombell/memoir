# Tonality

Tonality is a package that can convert musical keys between common notations in
DJing.

Supported notations:

- Camelot Key
- Open Key
- Musical
- Musical used by Beatport

## Usage

```go
key, err := tonality.ConvertKeyToNotation("4A", tonality.Musical)
if err != nil {
    // ...
}

// key => Fm
```

Tonality will detect the notation of the supplied key and convert to the given
notation.
