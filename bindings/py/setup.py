import pathlib
import shutil
from setuptools import setup
from setuptools.command.build_ext import build_ext
from setuptools.errors import CompileError
from setuptools.extension import Extension
from subprocess import CalledProcessError, check_call

HERE = pathlib.Path(__file__).parent
README = (HERE / "README.md").read_text()

class build_ext_external(build_ext):
    """Build the Go shared library and place it inside the package."""
    def build_extension(self, ext: Extension):
        library = pathlib.Path("src/minify/minify.so")
        if not library.exists():
            try:
                check_call(['go', 'build', '-buildmode=c-shared', '-o', str(library)])
            except CalledProcessError as e:
                raise CompileError('Go compilation failed!') from e
        target = pathlib.Path(self.build_lib) / "minify" / library.name
        target.parent.mkdir(parents=True, exist_ok=True)
        shutil.copyfile(library, target)

setup(
    name="tdewolff-minify",
    version="{VERSION}",
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
        "Programming Language :: Python :: 3.13",
        "Topic :: Internet :: WWW/HTTP",
        "Topic :: Software Development :: Pre-processors",
        "Topic :: Text Processing :: Markup",
    ],
    ext_modules=[
        Extension("minify", ["minify.go"]),
    ],
    cmdclass={"build_ext": build_ext_external},
    packages=["minify"],
    package_dir={"": "src"},
    include_package_data=True,
    cffi_modules=["build_minify.py:ffi"],
    install_requires=["cffi>=1.0.0"],
    zip_safe=False,
)
