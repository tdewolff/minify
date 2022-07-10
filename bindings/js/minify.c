#include <node_api.h>
#include <stdlib.h>
#include "minify.h"

napi_status get_string(napi_env env, napi_value v, char **s, size_t *size) {
    napi_status status;

    napi_valuetype type;
    status = napi_typeof(env, v, &type);
    if (status != napi_ok) {
        return status;
    } else if (type != napi_string) {
        return (napi_status)napi_string_expected;
    }

    size_t len, len_read;
    status = napi_get_value_string_utf8(env, v, NULL, 0, &len);
    if (status != napi_ok) {
        return status;
    }

    *s = (char *)malloc(len+1);
    status = napi_get_value_string_utf8(env, v, *s, len+1, &len_read);
    if (status != napi_ok) {
        free(*s);
        return status;
    } else if (len != len_read) {
        free(*s);
        return napi_generic_failure;
    }
    if (size != NULL) {
        *size = len;
    }
    return napi_ok;
}

void free_char_array(char **array, uint32_t length) {
    for (uint32_t i = 0; i < length; i++) {
        free(array[i]);
    }
    free(array);
}

napi_value config(napi_env env, napi_callback_info info) {
    napi_status status;

    size_t argc = 1;
    napi_value argv[1];
    status = napi_get_cb_info(env, info, &argc, argv, NULL, NULL);
    if (status != napi_ok) {
        return NULL;
    } else if (argc < 1) {
        napi_throw_type_error(env, NULL, "expected config argument");
        return NULL;
    }

    napi_valuetype type;
    status = napi_typeof(env, argv[0], &type);
    if (status != napi_ok) {
        return NULL;
    } else if (type != napi_object) {
        napi_throw_type_error(env, NULL, "config must be an object");
        return NULL;
    }

    napi_value properties;
    status = napi_get_all_property_names(env, argv[0], napi_key_own_only, napi_key_enumerable, napi_key_numbers_to_strings, &properties);
    if (status != napi_ok) {
        return NULL;
    }

    uint32_t length;
    status = napi_get_array_length(env, properties, &length);
    if (status != napi_ok) {
        return NULL;
    }

    char **keys = (char **)malloc(length * sizeof(char *));
    char **vals = (char **)malloc(length * sizeof(char *));
    for (uint32_t i = 0; i < length; i++) {
        napi_value nkey, nval;
        status = napi_get_element(env, properties, i, &nkey);
        if (status != napi_ok) {
            free_char_array(vals, i);
            free_char_array(keys, i);
            return NULL;
        }
        status = napi_get_property(env, argv[0], nkey, &nval);
        if (status != napi_ok) {
            free_char_array(vals, i);
            free_char_array(keys, i);
            return NULL;
        }

        status = get_string(env, nkey, &keys[i], NULL);
        if (status == napi_string_expected) {
            free_char_array(vals, i);
            free_char_array(keys, i);
            napi_throw_type_error(env, NULL, "config keys must be strings");
            return NULL;
        } else if (status != napi_ok) {
            free_char_array(vals, i);
            free_char_array(keys, i);
            return NULL;
        }

        status = napi_typeof(env, nval, &type);
        if (status != napi_ok) {
            free_char_array(vals, i);
            free_char_array(keys, i+1);
            return NULL;
        } else if (type == napi_boolean || type == napi_number) {
            status = napi_coerce_to_string(env, nval, &nval);
        } else if (type != napi_string) {
            free_char_array(vals, i);
            free_char_array(keys, i+1);
            napi_throw_type_error(env, NULL, "config values must be strings, integers, or booleans");
            return NULL;
        }

        status = get_string(env, nval, &vals[i], NULL);
        if (status != napi_ok) {
            free_char_array(vals, i);
            free_char_array(keys, i+1);
            return NULL;
        }
    }

    char *error = minifyConfig(keys, vals, length);
    free_char_array(vals, length);
    free_char_array(keys, length);
    if (error != NULL) {
        napi_throw_error(env, NULL, error);
        free(error);
        return NULL;
    }
    return NULL;
}

napi_value string(napi_env env, napi_callback_info info) {
    napi_status status;

    size_t argc = 2;
    napi_value argv[2];
    status = napi_get_cb_info(env, info, &argc, argv, NULL, NULL);
    if (status != napi_ok) {
        return NULL;
    } else if (argc < 2) {
        napi_throw_type_error(env, NULL, "expected mediatype and input arguments");
        return NULL;
    }

    char *mediatype, *input;
    size_t input_length;
    status = get_string(env, argv[0], &mediatype, NULL);
    if (status == napi_string_expected) {
        napi_throw_type_error(env, NULL, "mediatype must be a string");
        return NULL;
    } else if (status != napi_ok) {
        return NULL;
    }
    status = get_string(env, argv[1], &input, &input_length);
    if (status == napi_string_expected) {
        free(mediatype);
        napi_throw_type_error(env, NULL, "input must be a string");
        return NULL;
    } else if (status != napi_ok) {
        free(mediatype);
        return NULL;
    }

    char *output = (char *)malloc(input_length);
    long long output_length;

    char *error = minifyString(mediatype, input, input_length, output, &output_length);
    free(input);
    free(mediatype);
    if (error != NULL) {
        napi_throw_error(env, NULL, error);
        free(error);
        free(output);
        return NULL;
    }

    napi_value ret;
    status = napi_create_string_utf8(env, output, output_length, &ret);
    free(output);
    if (status != napi_ok) {
        return NULL;
    }
    return ret;
}

napi_value file(napi_env env, napi_callback_info info) {
    napi_status status;

    size_t argc = 3;
    napi_value argv[3];
    status = napi_get_cb_info(env, info, &argc, argv, NULL, NULL);
    if (status != napi_ok) {
        return NULL;
    } else if (argc < 3) {
        napi_throw_type_error(env, NULL, "expected mediatype, input, and output arguments");
        return NULL;
    }

    char *mediatype, *input, *output;
    status = get_string(env, argv[0], &mediatype, NULL);
    if (status == napi_string_expected) {
        napi_throw_type_error(env, NULL, "mediatype must be a string");
        return NULL;
    } else if (status != napi_ok) {
        return NULL;
    }
    status = get_string(env, argv[1], &input, NULL);
    if (status == napi_string_expected) {
        free(mediatype);
        napi_throw_type_error(env, NULL, "input must be a string");
        return NULL;
    } else if (status != napi_ok) {
        free(mediatype);
        return NULL;
    }
    status = get_string(env, argv[2], &output, NULL);
    if (status == napi_string_expected) {
        free(input);
        free(mediatype);
        napi_throw_type_error(env, NULL, "output must be a string");
        return NULL;
    } else if (status != napi_ok) {
        free(input);
        free(mediatype);
        return NULL;
    }

    char *error = minifyFile(mediatype, input, output);
    free(output);
    free(input);
    free(mediatype);
    if (error != NULL) {
        napi_throw_error(env, NULL, error);
        free(error);
        return NULL;
    }
    return NULL;
}

void cleanup() {
    minifyCleanup();
}

napi_value init(napi_env env, napi_value exports) {
    napi_status status;
    napi_value fnConfig, fnString, fnFile;

    status = napi_add_async_cleanup_hook(env, cleanup, NULL, NULL);

    status = napi_create_function(env, NULL, 0, config, NULL, &fnConfig);
    if (status != napi_ok) {
        return NULL;
    }
    status = napi_set_named_property(env, exports, "config", fnConfig);
    if (status != napi_ok) {
        return NULL;
    }

    status = napi_create_function(env, NULL, 0, string, NULL, &fnString);
    if (status != napi_ok) {
        return NULL;
    }
    status = napi_set_named_property(env, exports, "string", fnString);
    if (status != napi_ok) {
        return NULL;
    }

    status = napi_create_function(env, NULL, 0, file, NULL, &fnFile);
    if (status != napi_ok) {
        return NULL;
    }
    status = napi_set_named_property(env, exports, "file", fnFile);
    if (status != napi_ok) {
        return NULL;
    }
    return exports;
}

NAPI_MODULE(NODE_GYP_MODULE_NAME, init)
