create table if not exists Users (
	id integer primary key generated always as identity,
	nickname varchar not null,
	fullname varchar not null,
	email varchar (256) not null,
	about varchar
);

create table if not exists Forums (
	id integer primary key generated always as identity,
	title varchar not null,
	slug varchar not null,
	posts_cnt integer default 0 check (posts_cnt >= 0),
	threads_cnt integer default 0 check (threads_cnt >= 0),
	
	author_id integer references Users not null
);

create table if not exists Threads (
	id integer primary key generated always as identity,
	title varchar not null,
	message varchar not null,
	votes integer default 0,
	slug varchar,
	created_at timestamp default now(),
	
	forum_id integer references Forums not null,
	author_id integer references Users not null
);

create table if not exists Posts (
	id integer primary key generated always as identity,
	
	message varchar not null,
	edited bool default false,
	created_at timestamp default now(),
	
	parent_id integer references Posts default null,
	author_id integer references Users not null,
	thread_id integer references Threads not null
);

create unique index on Users (lower(nickname));
create unique index on Users (lower(email));

create unique index on Forums (lower(slug));