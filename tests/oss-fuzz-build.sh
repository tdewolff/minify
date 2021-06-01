# compile_go_fuzzer can be found in the oss-fuzz repository
root=$(pwd)
find parse/tests/* -maxdepth 0 -type d | while read target
do
    cd $root/$target
    fuzz_target=`echo $target | rev | cut -d'/' -f 1 | rev`
    compile_go_fuzzer . Fuzz parse-$fuzz_target gofuzz
done

find minify/tests/* -maxdepth 0 -type d | while read target
do
    cd $root/$target
    fuzz_target=`echo $target | rev | cut -d'/' -f 1 | rev`
    compile_go_fuzzer . Fuzz minify-$fuzz_target gofuzz
done
