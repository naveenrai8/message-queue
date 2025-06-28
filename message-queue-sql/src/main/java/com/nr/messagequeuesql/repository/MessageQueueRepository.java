package com.nr.messagequeuesql.repository;

import com.nr.messagequeuesql.Model.Message;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.UUID;

@Repository
public interface MessageQueueRepository extends JpaRepository<Message, UUID> {
}
