from distutils.core import setup
from Cython.Build import cythonize

setup(ext_modules=cythonize('printinfo.py'))


# 将py文件编译成so文件
# python setup.py build_ext --inplace
