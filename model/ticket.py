from pydantic import BaseModel
class FerryTicketData(BaseModel):
    title: str
    content: str
    priority: int
    category: str
    creator: str

