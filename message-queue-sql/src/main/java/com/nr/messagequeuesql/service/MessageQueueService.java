package com.nr.messagequeuesql.service;

import com.nr.messagequeuesql.Model.Message;
import com.nr.messagequeuesql.dto.MessageRequestDto;
import com.nr.messagequeuesql.dto.MessageResponseDto;
import com.nr.messagequeuesql.repository.MessageQueueRepository;
import jakarta.transaction.Transactional;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;
import java.util.UUID;

import static org.springframework.data.jpa.domain.AbstractPersistable_.id;

@Slf4j
@Service
@RequiredArgsConstructor
public class MessageQueueService {
    private final MessageQueueRepository repository;

    @Transactional
    public List<MessageResponseDto> getMessagesByCount(UUID clientId, Integer count, int leaseExpiredAtInSeconds) {
        List<MessageResponseDto> messages = new ArrayList<>();
        var res = this.repository.findNotAssignedNMessages(count, LocalDateTime.now());
        res.forEach(
                message ->
                {
                    LocalDateTime expiredAtInSeconds = LocalDateTime.now().plusSeconds(leaseExpiredAtInSeconds);
                    System.out.println(message.toString() + " " + expiredAtInSeconds);
                    this.repository.updateAssignedToAndLeaseExpiredAtTime(clientId, expiredAtInSeconds, message.getId());
                    messages.add(MessageResponseDto.builder()
                            .messageId(message.getId())
                            .messages(message.getMessage())
                            .leaseExpiredAt(expiredAtInSeconds)
                            .build());
                }
        );
        return messages;
    }

    public void addMessage(MessageRequestDto messageRequestDto) {
        var savedMessage = this.repository.save(Message.builder()
                .message(messageRequestDto.message())
                .createdAt(LocalDateTime.now())
                .build());
        log.info("Message saved successfully. {}", savedMessage);
    }

    @Transactional
    public void deleteMessage(UUID messageId, UUID clientId) {
        this.repository.deleteByIdForClient(messageId, clientId);
        log.info("Message deleted successfully. {}", id);
    }
}
