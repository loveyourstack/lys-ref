
export interface AuthorInput {
  name: string | undefined
}
export interface Author extends AuthorInput {
  id: number
  archived_at: Date | null
  archived_by_cascade: boolean
  book_count: number
  created_at: Date
  updated_at: Date
}
export function NewAuthor(): Author {
  return  {
    name: undefined,

    id: 0,
    archived_at: null,
    archived_by_cascade: false,
    book_count: 0,
    created_at: new Date(),
    updated_at: new Date(),
  }
}
export function GetAuthorInputFromItem(item: Author): AuthorInput {
  return  {
    name: item.name,
  }
}

// ------------------------------------------------------------------------------------------------------------------------------------------

export interface BookInput {
  author_fk: number | undefined
  name: string | undefined
}
export interface Book extends BookInput {
  id: number
  archived_at: Date | null
  archived_by_cascade: boolean
  author: string
  author_is_archived: boolean
  created_at: Date
  updated_at: Date
}
export function NewBook(): Book {
  return  {
    author_fk: undefined,
    name: undefined,

    id: 0,
    archived_at: null,
    archived_by_cascade: false,
    author: '',
    author_is_archived: false,
    created_at: new Date(),
    updated_at: new Date(),
  }
}
export function GetBookInputFromItem(item: Book): BookInput {
  return  {
    author_fk: item.author_fk,
    name: item.name,
  }
}