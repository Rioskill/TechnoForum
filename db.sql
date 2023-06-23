create table if not exists Users (
	id integer primary key generated always as identity,
	nickname varchar collate "C" not null,
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

create table if not exists ForumUserLinks (
	user_id integer references Users,
    forum_id integer references Forums,
    primary key(user_id, forum_id)
);

create table if not exists Threads (
	id integer primary key generated always as identity,
	title varchar not null,
	message varchar not null,
	votes_cnt integer default 0,
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
	path integer[] not null,
	parent_id integer references Posts default null,
	author_id integer references Users not null,
	thread_id integer references Threads not null
);

create table if not exists Vote(
    author_id integer references Users,
    thread_id integer references Threads,
    value int check (value = 1 or value = -1),
    primary key(thread_id, author_id)
);



create or replace function update_votes_cnt() returns trigger as $$
    begin
        if (tg_op = 'INSERT') then
            update Threads set votes_cnt = votes_cnt + NEW.value where id = NEW.thread_id;
            return NEW;
        elsif (tg_op = 'UPDATE') then
            update Threads set votes_cnt = votes_cnt - OLD.value + NEW.value where id = NEW.thread_id;
            return NEW;
        elsif (tg_op = 'DELETE') then
            update Threads set votes_cnt = votes_cnt - OLD.value where id = OLD.thread_id;
            return OLD;
        end if;
        return NULL;
    end;
$$ language plpgsql;

create or replace function update_thread_cnt() returns trigger as $$
    begin
        if (tg_op = 'INSERT') then
            update Forums set threads_cnt = threads_cnt + 1 where id = NEW.forum_id;
            return NEW;
        elsif (tg_op = 'DELETE') then
            update Forums set threads_cnt = threads_cnt - 1 where id = OLD.forum_id;
            return OLD;
        end if;
        return NULL;
    end;
$$ language plpgsql;

create or replace function on_post_insert() returns trigger as $$
    begin
        if (NEW.parent_id IS NULL) then
            NEW.path = array[NEW.id];
            return NEW;
        end if;

        if (NEW.thread_id != (SELECT thread_id FROM Posts WHERE id = NEW.parent_id)) then
            raise exception 'Thread id should be equal to one of parent' USING ERRCODE = '23000';
        end if;
        
        NEW.path = (SELECT path FROM Posts WHERE id = NEW.parent_id) || NEW.id;
        return NEW;
    end;
$$ language plpgsql;


create or replace function link_user_to_forum() returns trigger as $$
    begin
        insert into ForumUserLinks(user_id, forum_id)
            values (NEW.author_id, NEW.forum_id)
            on conflict do nothing;
        return NEW;
    end;
$$ language plpgsql;


create trigger on_vote
after insert or update or delete on Vote
    for each row execute procedure update_votes_cnt();

create trigger on_thread
after insert or delete on Threads
    for each row execute procedure update_thread_cnt();
   
create trigger before_post_insert
before insert on Posts
    for each row execute procedure on_post_insert();

create trigger link_user_to_forum_on_thread
after insert on Threads
    for each row execute procedure link_user_to_forum();

create unique index on Users (lower(nickname));
create unique index on Users (lower(email));
create unique index on Forums (lower(slug));
create unique index on Threads (lower(slug));

create unique index on Forums (lower(slug));

create index on posts (thread_id);
create index on posts ((path[1]));
create index on posts ((path[2:]));
