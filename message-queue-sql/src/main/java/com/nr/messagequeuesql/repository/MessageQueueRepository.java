package com.nr.messagequeuesql.repository;

import com.nr.messagequeuesql.Model.Message;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.stereotype.Repository;

import java.time.LocalDateTime;
import java.util.List;
import java.util.UUID;

@Repository
public interface MessageQueueRepository extends JpaRepository<Message, UUID> {

    @Query(
            value = "SELECT * from messages where assigned_to is null or lease_expired_at < :leaseExpiredAtTime limit :count for update skip locked",
            nativeQuery = true)
    List<Message> findNotAssignedNMessages(int count, LocalDateTime leaseExpiredAtTime);

    @Modifying
    @Query(
            value = " UPDATE messages set assigned_to = ?, lease_expired_at = ? where id = ?",
            nativeQuery = true
    )
    void updateAssignedToAndLeaseExpiredAtTime(UUID assignedTo, LocalDateTime leaseExpiredAtTime, UUID messageId);

    @Modifying
    @Query(
            value = "DELETE from messages where id = ? and assigned_to = ?",
            nativeQuery = true
    )
    void deleteByIdForClient(UUID messageId, UUID clientId);
}
