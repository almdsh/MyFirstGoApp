CREATE TABLE public.author(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL
);

CREATE TABLE public.book(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL
);

CREATE TABLE public.author_book(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    book_id UUID NOT NULL,
    author_id UUID ARRAY,

    CONSTRAINT fk_book FOREIGN KEY(book_id) REFERENCES book(id),
    CONSTRAINT fk_author FOREIGN KEY(author_id) REFERENCES author(id)
);

INSERT INTO book(name,auther_id) VALUES ('Book 1');
INSERT INTO book(name,auther_id) VALUES ('Book 1');
INSERT INTO book(name,auther_id) VALUES ('Book 1');