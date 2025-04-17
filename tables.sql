use wallet_signer;
drop table if exists address;
create table address
(
    id          int(20) not null AUTO_INCREMENT,
    chain_id    varchar(100),
    addr        varchar(100),
    private_key varchar(200),
    created_At  timestamp not null default now(),
    unique key (chain_id,addr),
    PRIMARY KEY (id)
);