import pathlib
from typing import Dict, Union

from ._ffi_minify import ffi

__all__ = ['MinifyError', 'config', 'file', 'string']


lib = ffi.dlopen(str(pathlib.Path(__file__).parent / 'minify.so'))


class MinifyError(ValueError):
    """
    Exception raised when error occurs during minification.
    """


def _check_error(cdata):
    if cdata:
        raise MinifyError(ffi.string(cdata).decode())


def config(configuration: Dict[str, Union[str,bool,int]]) -> None:
    """
    Configure minifier behavior.

    Supported configuration keys:
    * css-precision (int)
    * html-keep-comments (bool)
    * html-keep-conditional-comments (bool)
    * html-keep-default-attr-vals (bool)
    * html-keep-document-tags (bool)
    * html-keep-end-tags (bool)
    * html-keep-whitespace (bool)
    * html-keep-quotes (bool)
    * js-precision (int)
    * js-keep-var-names (bool)
    * js-version (int)
    * json-precision (int)
    * json-keep-numbers (bool)
    * svg-keep-comments (bool)
    * svg-precision (int)
    * xml-keep-whitespace (bool)
    """
    length = len(configuration)
    err_msg = lib.minifyConfig(
        [ffi.new('char[]', k.encode()) for k in configuration.keys()],
        [ffi.new('char[]', str(v).encode()) for v in configuration.values()],
        length)
    _check_error(err_msg)


def string(mediatype: str, string: str) -> str:
    """
    Minify code from a string.

    The minifier backend will be selected based on the specified MIME type.
    """
    input_bytes = string.encode()
    input_len = len(input_bytes)
    output_bytes = ffi.new(f'char[{input_len}]')
    output_len = ffi.new('long long *')
    err_msg = lib.minifyString(mediatype.encode(), input_bytes, input_len, output_bytes, output_len)
    _check_error(err_msg)
    return ffi.string(output_bytes, output_len[0]).decode()


def file(mediatype: str, input_filename: str, output_filename: str) -> None:
    """
    Minify a file and write output to another file.

    The minifier backend will be selected based on the specified MIME type.
    """
    err_msg = lib.minifyFile(mediatype.encode(), input_filename.encode(), output_filename.encode())
    _check_error(err_msg)
