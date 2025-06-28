package com.nr.messagequeuesql.service;

import com.nr.messagequeuesql.Model.Message;
import com.nr.messagequeuesql.dto.MessageRequestDto;
import com.nr.messagequeuesql.dto.MessageResponseDto;
import com.nr.messagequeuesql.repository.MessageQueueRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.util.UUID;

@Slf4j
@Service
@RequiredArgsConstructor
public class MessageQueueService {
    private final MessageQueueRepository repository;

    public MessageResponseDto getMessagesByCount(Integer count) {
        return MessageResponseDto.builder()
                .messages(this.repository.findAll().stream().map(
                        Message::getMessage
                ).toList())
                .build();
    }

    public void addMessage(MessageRequestDto messageRequestDto) {
        var savedMessage = this.repository.save(Message.builder()
                .message(messageRequestDto.message())
                .build());
        log.info("Message saved successfully. {}", savedMessage);
    }

    public void deleteMessage(String id) {
        this.repository.deleteById(UUID.fromString(id));
        log.info("Message deleted successfully. {}", id);
    }
}
