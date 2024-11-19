# Bytocol

Enabling easy creation of byte-protocols for transmission over TCP connections.

Essentially this is a variation on the Protobuf idea, for simpler
protocols that want to leverage TCP instead of carrying around the whole `protoc`
and gRPC services.

It works based on struct tags much like JSON or any other encoding format, but
instead produces byte-ordered packets for transmission over `io.Reader` and
`io.Writer` interfaces.

As of this first version, Go reflection is used for the struct tags, there could
be benefits later of doing a protoc type situation in which Go generate is
leveraged to build the encoders at build-time.

## Usage

### Message Interface

The bytocol package will consume any structs that implement the `bytocol.Message`
interface. This interface requires the `BytocolMessage() bytocol.MessageInfo` method
be implemented. This method returns a struct instructing the encoder on the
top-level information about the message, such as a debug name and the type indicator.

Type-indicators are commonly a single byte number indicating the message type so
that any further refactors of names do not interfere with the protocol. It is
recommended to keep a list of them as constants in your code base so that they
are not confused or overlapping. __The value 0 is reserved as an error message
type indicator.__

### Field Tags

Within the struct that implements, you can now use the `bytocol:".."` struct tags
on fields that are part of the message. Only fields using this tag will be used.

The tags follow the pattern `order number, options...` where the options portion
is optional list of encoding options. The first portion is an unsigned integer
indicating the field order for encoding.

#### Available Options

The following is a table of available options, most of these are tags but for
options that accept values they are key-value pairs as `key=value`. When using
multiple options they are comma separated.

| Option Key | Description | Accepts Value | Value Type |
|:-----------|:------------|:-------------:|:-----------|
| `null-terminated` | String is encoded as null terminated | No |   |
| `length-prefix`   | Bit-size of length preceded this value | Yes | 8, 16, 32, 64 |

### Data Types

Most primitive types are encoded with reasonable defaults based on their type,
strings are the trickier ones since they are encoded as byte data blobs and as
such need some way to indicate their length. For strings, it is recommended to
use the `length-prefix` option with a reasonable size for the maximum length.

All numerical types are encoded as Big-Endian bytes up to their data size. For
instance, a signed 32-bit integer will be encoded as 4 bytes. To prevent platform differences, the `int` and `uint` types are transmitted as 64-bit values.

Booleans are single byte values for transmission sake.

Any interfaces, structs, maps, slices, and arrays are transmitted using the `gob.Encode`
encoder. This may change in the future, but this makes it easier for this first
version.

Standard `error` types can be wrapped with the provided `bytocol.Error` function
which converts it into `bytcol.ErrorMessage` for transmission. This type uses
reserved type-indicator 0 and contains the message as a 16-bit length-prefixed
string. This enforces a maximum error message length of 65,535 bytes. If you
just want to create a new error directly the helper `bytocol.NewError` is
provided as well which takes a string argument.
