#ifndef __WRAPPER_H_
#define __WRAPPER_H_

#include <stdint.h>

typedef char *(*fn_malloc)(uint64_t size);

#ifdef _MSC_VER
    #define __SIZE_TYPE__ void*
    #define _Bool char
    #define _Complex
    #include "pyeoskit.h"
#else
    #include "libpyeoskit.h"
#endif

#endif//__WRAPPER_H_
