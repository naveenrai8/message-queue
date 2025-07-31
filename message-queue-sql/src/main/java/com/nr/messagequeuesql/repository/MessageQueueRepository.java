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
            value = """
    (
      SELECT *
      FROM messages
      WHERE assigned_to IS NULL
      LIMIT :count
      FOR UPDATE SKIP LOCKED
    )
    UNION ALL
    (
      SELECT *
      FROM messages
      WHERE lease_expired_at < :currentDateTime
      LIMIT :count
      FOR UPDATE SKIP LOCKED
    )
    LIMIT :count;
    
""", nativeQuery = true
    )
    List<Message> findAvailableMessages(int count, LocalDateTime currentDateTime);

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
//    @Query(
//            value = "UPDATE messages set lease_expired_at = FROM_UNIXTIME(2147483600)  where id = ? and assigned_to = ?",
//            nativeQuery = true
//    )
    void deleteByIdForClient(UUID messageId, UUID clientId);
}
