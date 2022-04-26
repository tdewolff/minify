rm -rf dist
rm -rf tdewolff_minify.egg-info
go get -u all
python -m build --sdist
twine upload dist/*
