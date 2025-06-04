-- users: пользователи системы
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email TEXT UNIQUE NOT NULL,
  name TEXT NOT NULL,
  role TEXT NOT NULL,          -- operator, supervisor, admin
  password_hash TEXT NOT NULL, -- пароль в захешированном виде
  created_at TIMESTAMPTZ DEFAULT now()
);

-- projects: проекты под супервайзером
CREATE TABLE projects (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  supervisor_id UUID REFERENCES users(id)
);

-- chats: чаты проекта
CREATE TABLE chats (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id UUID REFERENCES projects(id),
  name TEXT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT now()
);

-- chat_participants: кто участвует в чате
CREATE TABLE chat_participants (
  chat_id UUID REFERENCES chats(id),
  user_id UUID REFERENCES users(id),
  joined_at TIMESTAMPTZ DEFAULT now(),
  PRIMARY KEY(chat_id, user_id)
);

-- messages: сообщения в чатах
CREATE TABLE messages (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chat_id UUID REFERENCES chats(id),
  sender_id UUID REFERENCES users(id),
  content TEXT,
  attachment_url TEXT,
  sent_at TIMESTAMPTZ DEFAULT now(),
  read BOOLEAN DEFAULT FALSE
);

-- schedule: расписание операторов
CREATE TABLE schedule (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id),
  date DATE NOT NULL,
  shift TEXT NOT NULL,         -- например \"09:00-17:00\"
  comment TEXT
);

-- ipr: индивидуальные планы развития
CREATE TABLE ipr (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id),
  goals TEXT NOT NULL,
  deadline DATE,
  status TEXT DEFAULT 'in_progress', -- in_progress, done
  comments TEXT,
  created_at TIMESTAMPTZ DEFAULT now()
);

-- documents: загруженные файлы
CREATE TABLE documents (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id UUID REFERENCES projects(id),
  title TEXT NOT NULL,
  file_url TEXT NOT NULL,
  type TEXT,                   -- регламент, инструкция и т.п.
  uploaded_by UUID REFERENCES users(id),
  uploaded_at TIMESTAMPTZ DEFAULT now()
);

-- reports: отчёты супервайзера
CREATE TABLE reports (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id UUID REFERENCES projects(id),
  author_id UUID REFERENCES users(id),
  content TEXT,
  created_at TIMESTAMPTZ DEFAULT now()
);