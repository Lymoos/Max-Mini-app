-- описание задачи и список внутри: «Купить в аптеке» -> лекарства по рецепту
ALTER TABLE tasks ADD COLUMN note TEXT NOT NULL DEFAULT '';

CREATE TABLE task_items (
    id          SERIAL PRIMARY KEY,
    task_id     INT NOT NULL REFERENCES tasks (id) ON DELETE CASCADE,
    title       TEXT NOT NULL,
    medicine_id TEXT NOT NULL DEFAULT '',
    done        BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX task_items_task_idx ON task_items (task_id);
