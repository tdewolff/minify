import pathlib
from subprocess import call
from setuptools import Extension, setup
from setuptools.command.build_ext import build_ext
from setuptools.errors import CompileError

HERE = pathlib.Path(__file__).parent
README = (HERE / "README.md").read_text()

class build_go_ext(build_ext):
    """Custom command to build extension from Go source files"""
    def build_extension(self, ext):
        ext_path = self.get_ext_fullpath(ext.name)
        cmd = ['go', 'build', '-buildmode=c-shared', '-o', ext_path]
        #cmd += ext.sources
        out = call(cmd)
        if out != 0:
            raise CompileError('Go build failed')

def get_version():
    with open('go.mod') as f:
        for line in f:
            line = line.strip()
            if line.startswith('github.com/tdewolff/minify/v2'):
                return line.split()[1][1:]
    raise CompileError('Version retrieval failed')

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
    ext_modules=[
        Extension("minify", ["minify.go"]),
    ],
    cmdclass={"build_ext": build_go_ext},
    zip_safe=False,
)
