import cffi

C_DEFS = """
char * minifyConfig(char **ckeys, char **cvals, long long length);
char * minifyString(char *cmediatype, char *cinput, long long input_length, char *coutput, long long *output_length);
char * minifyFile(char *cmediatype, char *cinput, char *coutput);
"""

ffi = cffi.FFI()
ffi.set_source('minify._ffi_minify', None)
ffi.cdef(C_DEFS)

if __name__ == '__main__':
    ffi.compile()
