alter table request add column movement_type_id integer references movements_type(id);

insert into movements_type (id, name, description) values
(1, 'Entrada', 'Movimento de entrada de produtos no estoque'),
(2, 'Saída', 'Movimento de saída de produtos do estoque');