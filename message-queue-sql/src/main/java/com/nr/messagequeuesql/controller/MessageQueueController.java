package com.nr.messagequeuesql.controller;

import com.nr.messagequeuesql.dto.MessageRequestDto;
import com.nr.messagequeuesql.dto.MessageResponseDto;
import com.nr.messagequeuesql.service.MessageQueueService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

@RestController
@RequestMapping("/messages")
@RequiredArgsConstructor
@Slf4j
public class MessageQueueController {

    private final MessageQueueService service;

    @Value("${app.leaseExpiredAt:10}")
    private int LEASE_EXPIRED_DEFAULT_TIME;

    @GetMapping
    public ResponseEntity<List<MessageResponseDto>> getMessages(
            @RequestParam() UUID clientId,
            @RequestParam(required = false) Integer count,
            @RequestParam(required = false) Optional<Integer> leaseExpiredAtInSeconds) {
        log.info("Received request to get all messages");
        if (count == null) {
            count = 1;
        }

        return ResponseEntity.ok(this.service.getMessagesByCount(clientId, count, leaseExpiredAtInSeconds.orElse(LEASE_EXPIRED_DEFAULT_TIME)));
    }

    @PostMapping
    public ResponseEntity<Void> addMessage(@RequestBody MessageRequestDto messageRequestDto) {
        this.service.addMessage(messageRequestDto);
        return ResponseEntity.ok().build();
    }

    @DeleteMapping
    public ResponseEntity<Void> deleteMessage(@RequestParam UUID messageId, @RequestParam UUID clientId) {
        this.service.deleteMessage(messageId, clientId);
        return ResponseEntity.ok().build();
    }
}
