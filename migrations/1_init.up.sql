CREATE TABLE PEOPLES
(
    id bigserial NOT NULL,
	name text NOT NULL,
    surname text NOT NULL,
    patronymic text,
    PRIMARY KEY (id)
);

CREATE TABLE CARS
(
    id bigserial NOT NULL,
	reg_num text NOT NULL,
    mark text NOT NULL,
    model text NOT NULL,
    year integer,
    owner_id integer,
    PRIMARY KEY (id),
    FOREIGN KEY (owner_id) REFERENCES PEOPLES(id)
);