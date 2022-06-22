#include <Python.h>

char *minifyConfig(char **, char **, long long);
char *minifyString(char *, char *, long long, char *, long long *);
char *minifyFile(char *, char *, char *);

PyObject *config(PyObject *self, PyObject *args) {
    PyObject *pyconfig;
    if (PyArg_ParseTuple(args, "O", &pyconfig) == 0 || !PyDict_Check(pyconfig)) {
        PyErr_SetString(PyExc_ValueError, "expected config argument");
        return NULL;
    } else if (!PyDict_Check(pyconfig)) {
        PyErr_SetString(PyExc_ValueError, "config must be a dict[str,str|bool|int]");
        return NULL;
    }

    Py_ssize_t length = PyDict_Size(pyconfig);
    const char **keys = (const char **)malloc(length * sizeof(const char *));
    const char **vals = (const char **)malloc(length * sizeof(const char *));

    Py_ssize_t pos = 0;
    PyObject *pykey, *pyval;
    while (PyDict_Next(pyconfig, &pos, &pykey, &pyval)) {
        const char *key = PyUnicode_AsUTF8(pykey); // handles deallocation
        if (key == NULL) {
            PyErr_SetString(PyExc_ValueError, "config must be a dict[str,str|bool|int]");
            free(vals);
            free(keys);
            return NULL;
        }
        keys[pos-1] = key;

        int decref = 0;
        if (PyBool_Check(pyval) || PyLong_Check(pyval)) {
            pyval = PyObject_Str(pyval);
            if (pyval == NULL) {
                PyErr_SetString(PyExc_ValueError, "config must be a dict[str,str|bool|int]");
                free(vals);
                free(keys);
                return NULL;
            }
            decref = 1;
        }
        const char *val = PyUnicode_AsUTF8(pyval); // handles deallocation
        if (val == NULL) {
            PyErr_SetString(PyExc_ValueError, "config must be a dict[str,str|bool|int]");
            free(vals);
            free(keys);
            return NULL;
        }
        vals[pos-1] = val;

        if (decref == 1) {
            Py_DECREF(pyval);
        }
    }

    char *error = minifyConfig((char **)keys, (char **)vals, length);
    free(vals);
    free(keys);
    if (error != NULL) {
        PyErr_SetString(PyExc_ValueError, error);
        free(error);
        return NULL;
    }
    Py_RETURN_NONE;
}

PyObject *string(PyObject *self, PyObject *args) {
    PyObject *pymediatype, *pyinput;
    if (PyArg_ParseTuple(args, "OO", &pymediatype, &pyinput) == 0) {
        PyErr_SetString(PyExc_ValueError, "expected mediatype and input arguments");
        return NULL;
    }

    const char *mediatype = PyUnicode_AsUTF8(pymediatype); // handles deallocation
    if (mediatype == NULL) {
        PyErr_SetString(PyExc_ValueError, "mediatype must be a string");
        return NULL;
    }

    Py_ssize_t input_length; // not including trailing NULL-byte
    const char *input = PyUnicode_AsUTF8AndSize(pyinput, &input_length); // handles deallocation
    if (input == NULL) {
        PyErr_SetString(PyExc_ValueError, "input must be a string");
        return NULL;
    }

    long long output_length; // not including trailing NULL-byte
    char *output = (char *)malloc(input_length);
    char *error = minifyString((char *)mediatype, (char *)input, (long long)input_length, output, &output_length);
    if (error != NULL) {
        PyErr_SetString(PyExc_ValueError, error);
        free(error);
        return NULL;
    }

    PyObject *pyoutput = PyUnicode_DecodeUTF8(output, (Py_ssize_t)output_length, NULL);
    free(output);
    return pyoutput;
}

PyObject *file(PyObject *self, PyObject *args) {
    PyObject *pymediatype, *pyinput, *pyoutput;
    if (PyArg_ParseTuple(args, "OOO", &pymediatype, &pyinput, &pyoutput) == 0) {
        PyErr_SetString(PyExc_ValueError, "expected mediatype, input, and output arguments");
        return NULL;
    }

    const char *mediatype = PyUnicode_AsUTF8(pymediatype); // handles deallocation
    if (mediatype == NULL) {
        return NULL;
    }

    const char *input = PyUnicode_AsUTF8(pyinput); // handles deallocation
    if (input == NULL) {
        return NULL;
    }

    const char *output = PyUnicode_AsUTF8(pyoutput); // handles deallocation
    if (output == NULL) {
        return NULL;
    }

    char *error = minifyFile((char *)mediatype, (char *)input, (char *)output);
    if (error != NULL) {
        PyErr_SetString(PyExc_ValueError, error);
        free(error);
        return NULL;
    }
    Py_RETURN_NONE;
}

static PyMethodDef MinifyMethods[] = {
    {"config", config, METH_VARARGS, "Configure minify options."},
    {"string", string, METH_VARARGS, "Minify string."},
    {"file", file, METH_VARARGS, "Minify file."},
    {NULL, NULL, 0, NULL}
};

static struct PyModuleDef minifymodule = {
   PyModuleDef_HEAD_INIT, "minify", NULL, -1, MinifyMethods
};

PyMODINIT_FUNC
PyInit_minify(void)
{
    return PyModule_Create(&minifymodule);
}
