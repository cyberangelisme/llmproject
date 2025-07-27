from utils import *

from glob import glob
from langchain_community.vectorstores.chroma import Chroma
from langchain_community.document_loaders import TextLoader,CSVLoader,PyMuPDFLoader
from langchain.text_splitter import RecursiveCharacterTextSplitter

def doc2vec():
    # 定义文档分割器
    text_splitter = RecursiveCharacterTextSplitter(chunk_size=300, chunk_overlap=50)
    # 当前文件data_process.py所在目录 + data目录
    dir_path = os.path.join(os.path.dirname(__file__), './data/input/')

    # 找文件
    documents = []

    for file_path in glob(dir_path + '*.*'):
        if file_path.endswith('.txt'):
            loader = TextLoader(file_path)
        elif file_path.endswith('.csv'):
            loader = CSVLoader(file_path)
        elif file_path.endswith('.pdf'):
            loader = PyMuPDFLoader(file_path)
        else:
            raise ValueError("不支持的文件格式")
        if loader:
            documents += loader.load_and_split(text_splitter)
    print(documents)
   


    if documents:
        vdb = Chroma.from_documents(
            documents=documents,
            embedding=get_embeddings_model(),
            persist_directory=os.path.join(os.path.dirname(__file__),'./data/db/')
        )
        vdb.persist()
    
   

if  __name__ == '__main__':
    doc2vec()