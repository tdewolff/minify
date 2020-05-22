function compile_fuzzer {
  path=$1
  function=$2
  fuzzer=$3

  go-fuzz -func $function -o $fuzzer.a $path

  $CXX $CXXFLAGS $LIB_FUZZING_ENGINE $fuzzer.a -o $OUT/$fuzzer
}

find $GOPATH/src/github.com/tdewolff/minify/tests/*/ -maxdepth 1 -type d | while read target
do
    fuzz_target=`echo $target | cut -d'/' -f 9`
    compile_fuzzer github.com/tdewolff/minify/tests/$fuzz_target Fuzz $fuzz_target-fuzzer
done
