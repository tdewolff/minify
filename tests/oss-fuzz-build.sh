# compile_go_fuzzer can be found in the oss-fuzz repository
root=$(pwd)
find parse/tests/* -maxdepth 0 -type d | while read target
do
    cd $root/$d
    fuzz_target=`echo $target | rev | cut -d'/' -f 1 | rev`
    compile_go_fuzzer github.com/tdewolff/parse/tests/$fuzz_target Fuzz parse-$fuzz_target-fuzzer
done

find minify/tests/* -maxdepth 0 -type d | while read target
do
    cd $root/$d
    fuzz_target=`echo $target | rev | cut -d'/' -f 1 | rev`
    compile_go_fuzzer github.com/tdewolff/minify/tests/$fuzz_target Fuzz minify-$fuzz_target-fuzzer
done
