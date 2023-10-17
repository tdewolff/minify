import pathlib
from setuptools import setup, Distribution
from setuptools.errors import CompileError


HERE = pathlib.Path(__file__).parent
README = (HERE / "README.md").read_text()


def get_version():
    with open('go.mod') as f:
        for line in f:
            line = line.strip()
            if line.startswith('github.com/tdewolff/minify/v2'):
                return line.split()[1][1:]
    raise CompileError('Version retrieval failed')


class BinaryDistribution(Distribution):
    """
    Distribution which always forces binary packages to be platform-specific.

    Based on https://stackoverflow.com/a/36886459/1795505.
    """
    def has_ext_modules(self):
        return True


setup(
    name="tdewolff-minify",
    version=get_version(),
    description="Go minifiers for web formats",
    long_description=README,
    long_description_content_type="text/markdown",
    url="https://github.com/tdewolff/minify",
    author="Taco de Wolff",
    author_email="tacodewolff@gmail.com",
    license="MIT",
    classifiers = [
        "Development Status :: 4 - Beta",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
        "Programming Language :: JavaScript",
        "Programming Language :: Python",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3 :: Only",
        "Programming Language :: Python :: 3.7",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: Python :: 3.12",
        "Topic :: Internet :: WWW/HTTP",
        "Topic :: Software Development :: Pre-processors",
        "Topic :: Text Processing :: Markup",
    ],
    packages=["minify"],
    package_dir={"": "src"},
    package_data={"minify": ["_minify.so"]},
    setup_requires=["cffi>=1.0.0"],
    cffi_modules=["build_minify.py:ffi"],
    install_requires=["cffi>=1.0.0"],
    zip_safe=False,
    distclass=BinaryDistribution,
)
