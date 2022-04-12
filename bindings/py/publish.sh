rm -rf dist
rm -rf tdewolff_minify.egg-info
python -m build --sdist
twine upload dist/*
